package fbhttp

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/filebrowser/filebrowser/v2/auth"
	"github.com/filebrowser/filebrowser/v2/requestinfo"
	"github.com/golang-jwt/jwt/v5"
	"github.com/golang-jwt/jwt/v5/request"

	fberrors "github.com/filebrowser/filebrowser/v2/errors"
	"github.com/filebrowser/filebrowser/v2/storage"
	"github.com/filebrowser/filebrowser/v2/users"
)

const (
	DefaultTokenExpirationTime = time.Hour * 2

	maxAuthBodySize = 1 << 20 // 1 MiB
)

type userInfo struct {
	ID                    uint              `json:"id"`
	Locale                string            `json:"locale"`
	ViewMode              users.ViewMode    `json:"viewMode"`
	SingleClick           bool              `json:"singleClick"`
	RedirectAfterCopyMove bool              `json:"redirectAfterCopyMove"`
	Perm                  users.Permissions `json:"perm"`
	Commands              []string          `json:"commands"`
	LockPassword          bool              `json:"lockPassword"`
	HideDotfiles          bool              `json:"hideDotfiles"`
	DateFormat            bool              `json:"dateFormat"`
	Username              string            `json:"username"`
	AceEditorTheme        string            `json:"aceEditorTheme"`
}

type authToken struct {
	User          userInfo `json:"user"`
	PasswordStamp string   `json:"pwd"`
	jwt.RegisteredClaims
}

// Bind sessions to the current password without exposing its stored hash.
func passwordStamp(key []byte, password string) string {
	mac := hmac.New(sha256.New, key)
	_, _ = mac.Write([]byte(password))
	return hex.EncodeToString(mac.Sum(nil))
}

type extractor []string

func (e extractor) ExtractToken(r *http.Request) (string, error) {
	token, _ := request.HeaderExtractor{"X-Auth"}.ExtractToken(r)

	// Checks if the token isn't empty and if it contains two dots.
	// The former prevents incompatibility with URLs that previously
	// used basic auth.
	if token != "" && strings.Count(token, ".") == 2 {
		return token, nil
	}

	if r.Method == http.MethodGet {
		cookie, _ := r.Cookie("auth")
		if cookie != nil && strings.Count(cookie.Value, ".") == 2 {
			return cookie.Value, nil
		}
	}

	return "", request.ErrNoTokenInRequest
}

func withUser(fn handleFunc) handleFunc {
	return withSession(fn, false)
}

func withSession(fn handleFunc, allowRotated bool) handleFunc {
	return func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
		keyFunc := func(_ *jwt.Token) (interface{}, error) {
			return d.settings.Key, nil
		}

		var tk authToken
		p := jwt.NewParser(jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}), jwt.WithExpirationRequired())
		token, err := request.ParseFromRequest(r, &extractor{}, keyFunc, request.WithClaims(&tk), request.WithParser(p))
		if err != nil || token == nil || !token.Valid {
			return http.StatusUnauthorized, nil
		}

		var valid bool
		if allowRotated {
			valid, err = d.store.Sessions.ValidFamily(tk.ID, tk.User.ID, time.Now())
		} else {
			valid, err = d.store.Sessions.Valid(tk.ID, tk.User.ID, time.Now())
		}
		if err != nil {
			return http.StatusInternalServerError, err
		}
		if !valid {
			return http.StatusUnauthorized, nil
		}
		d.sessionID = tk.ID

		expiresSoon := tk.ExpiresAt != nil && time.Until(tk.ExpiresAt.Time) < time.Hour
		updated := tk.IssuedAt != nil && tk.IssuedAt.Unix() < d.store.Users.LastUpdate(tk.User.ID)

		if expiresSoon || updated {
			w.Header().Add("X-Renew-Token", "true")
		}

		d.user, err = d.store.Users.Get(d.server.Root, d.server.FollowExternalSymlinks, tk.User.ID)
		if err != nil {
			return http.StatusUnauthorized, nil
		}
		if !hmac.Equal([]byte(tk.PasswordStamp), []byte(passwordStamp(d.settings.Key, d.user.Password))) {
			return http.StatusUnauthorized, nil
		}

		canonicalizeRequestPath(r)
		return fn(w, r, d)
	}
}

func withAdmin(fn handleFunc) handleFunc {
	return withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
		if !d.user.Perm.Admin {
			return http.StatusForbidden, nil
		}

		return fn(w, r, d)
	})
}

func loginHandler(tokenExpireTime time.Duration) handleFunc {
	return func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
		if r.Body != nil {
			r.Body = http.MaxBytesReader(w, r.Body, maxAuthBodySize)
		}

		auther, err := d.store.Auth.Get(d.settings.AuthMethod)
		if err != nil {
			return http.StatusInternalServerError, err
		}

		user, err := auther.Auth(r, d.store.Users, d.settings, d.server)
		switch {
		case errors.Is(err, fberrors.ErrTooManyAttempts):
			setPasswordRetryAfter(w, err)
			return http.StatusTooManyRequests, nil
		case errors.Is(err, os.ErrPermission):
			return http.StatusForbidden, nil
		case err != nil:
			return http.StatusInternalServerError, err
		}

		return printToken(w, r, d, user, tokenExpireTime)
	}
}

func setPasswordRetryAfter(w http.ResponseWriter, err error) {
	seconds := 6
	var limited *auth.AttemptLimitError
	if errors.As(err, &limited) {
		seconds = max(1, int((limited.RetryAfter+time.Second-1)/time.Second))
	}
	w.Header().Set("Retry-After", strconv.Itoa(seconds))
}

type signupBody struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

var signupHandler = func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	if !d.settings.Signup {
		return http.StatusMethodNotAllowed, nil
	}

	if r.Body == nil {
		return http.StatusBadRequest, nil
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxAuthBodySize)

	info := &signupBody{}
	err := json.NewDecoder(r.Body).Decode(info)
	if err != nil {
		return http.StatusBadRequest, err
	}

	if info.Password == "" || info.Username == "" {
		return http.StatusBadRequest, nil
	}

	user := &users.User{
		Username: info.Username,
	}

	d.settings.Defaults.Apply(user)

	// Users signed up via the signup handler should never become admins, even
	// if that is the default permission.
	user.Perm.Admin = false

	// Self-registered users should not inherit execution capabilities from
	// default settings, regardless of what the administrator has configured
	// as the default. Execution rights must be explicitly granted by an admin.
	user.Perm.Execute = false
	user.Commands = []string{}

	pwd, err := users.ValidateAndHashPwd(info.Password, d.settings.MinimumPasswordLength)
	if err != nil {
		return http.StatusBadRequest, err
	}

	user.Password = pwd

	err = d.settings.CreateSignupHome(user, d.server.Root)
	if err != nil {
		return errToStatus(err), err
	}

	log.Printf("new user: %s, home dir: [%s].", user.Username, user.Scope)

	err = d.store.Users.SaveProvisioned(user, true)
	if errors.Is(err, fberrors.ErrExist) {
		return http.StatusConflict, err
	} else if err != nil {
		return http.StatusInternalServerError, err
	}

	return http.StatusOK, nil
}

func renewHandler(tokenExpireTime time.Duration) handleFunc {
	return withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
		w.Header().Set("X-Renew-Token", "false")
		return issueToken(w, r, d, d.user, tokenExpireTime, d.sessionID)
	})
}

func printToken(w http.ResponseWriter, r *http.Request, d *data, user *users.User, tokenExpirationTime time.Duration) (int, error) {
	return issueToken(w, r, d, user, tokenExpirationTime, "")
}

func issueToken(w http.ResponseWriter, r *http.Request, d *data, user *users.User, tokenExpirationTime time.Duration, previous string) (int, error) {
	claims := &authToken{
		PasswordStamp: passwordStamp(d.settings.Key, user.Password),
		User: userInfo{
			ID:                    user.ID,
			Locale:                user.Locale,
			ViewMode:              user.ViewMode,
			SingleClick:           user.SingleClick,
			RedirectAfterCopyMove: user.RedirectAfterCopyMove,
			Perm:                  user.Perm,
			LockPassword:          user.LockPassword,
			Commands:              user.Commands,
			HideDotfiles:          user.HideDotfiles,
			DateFormat:            user.DateFormat,
			Username:              user.Username,
			AceEditorTheme:        user.AceEditorTheme,
		},
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenExpirationTime)),
			Issuer:    "File Browser",
			ID:        rand.Text(),
		},
	}

	if previous != "" {
		family, _, _ := strings.Cut(previous, ".")
		claims.ID = family + "." + rand.Text()
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(d.settings.Key)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	if previous == "" {
		err = d.store.Sessions.Create(claims.ID, user.ID, claims.ExpiresAt.Time)
	} else {
		err = d.store.Sessions.Replace(previous, claims.ID, user.ID, claims.ExpiresAt.Time)
	}
	if errors.Is(err, storage.ErrInvalidSession) {
		return http.StatusUnauthorized, nil
	}
	if err != nil {
		return http.StatusInternalServerError, err
	}
	setAuthCookie(w, r, d, signed, tokenExpirationTime)

	w.Header().Set("Content-Type", "text/plain")
	if _, err := w.Write([]byte(signed)); err != nil {
		return http.StatusInternalServerError, err
	}
	return 0, nil
}

func setAuthCookie(w http.ResponseWriter, r *http.Request, d *data, token string, lifetime time.Duration) {
	path := strings.TrimRight(d.server.BaseURL, "/") + "/"
	maxAge := int(lifetime.Seconds())
	if token == "" {
		maxAge = -1
	}
	http.SetCookie(w, &http.Cookie{Name: "auth", Value: token, Path: path, HttpOnly: true, Secure: requestinfo.Resolve(r, d.server.TrustedProxies).HTTPS, SameSite: http.SameSiteStrictMode, MaxAge: maxAge})
}

var logoutHandler = withSession(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	if err := d.store.Sessions.Delete(d.sessionID); err != nil {
		return http.StatusInternalServerError, err
	}
	setAuthCookie(w, r, d, "", 0)
	return http.StatusNoContent, nil
}, true)
