package fbhttp

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/filebrowser/filebrowser/v2/requestinfo"
	"github.com/filebrowser/filebrowser/v2/share"
	"golang.org/x/crypto/bcrypt"
)

const shareAccessTTL = time.Hour

func shareAccessValue(key []byte, link *share.Link, expires int64) string {
	mac := hmac.New(sha256.New, key)
	_, _ = fmt.Fprintf(mac, "%s\x00%s\x00%s\x00%d", link.Hash, link.PasswordHash, link.Token, expires)
	return strconv.FormatInt(expires, 10) + "." + hex.EncodeToString(mac.Sum(nil))
}

func authenticateShareRequest(w http.ResponseWriter, r *http.Request, d *data, link *share.Link) (int, error) {
	if link.PasswordHash == "" {
		return 0, nil
	}
	name := "fb_share_" + link.Hash
	password, err := url.QueryUnescape(r.Header.Get("X-SHARE-PASSWORD"))
	if err != nil {
		return http.StatusBadRequest, err
	}
	if password == "" {
		cookie, err := r.Cookie(name)
		if err == nil {
			timestamp, _, _ := strings.Cut(cookie.Value, ".")
			expires, err := strconv.ParseInt(timestamp, 10, 64)
			if err == nil && expires > time.Now().Unix() && (link.Expire == 0 || link.Expire > time.Now().Unix()) && hmac.Equal([]byte(cookie.Value), []byte(shareAccessValue(d.settings.Key, link, expires))) {
				return 0, nil
			}
		}
		// A permanent URL token must never substitute for the password.
		return http.StatusUnauthorized, nil
	}
	if err := bcrypt.CompareHashAndPassword([]byte(link.PasswordHash), []byte(password)); err != nil {
		return http.StatusUnauthorized, nil
	}
	expires := time.Now().Add(shareAccessTTL).Unix()
	if link.Expire != 0 && link.Expire < expires {
		expires = link.Expire
	}
	path := strings.TrimRight(d.server.BaseURL, "/") + "/api/public/"
	http.SetCookie(w, &http.Cookie{Name: name, Value: shareAccessValue(d.settings.Key, link, expires), Path: path, HttpOnly: true, Secure: requestinfo.Resolve(r, d.server.TrustedProxies).HTTPS, SameSite: http.SameSiteStrictMode, MaxAge: int(expires - time.Now().Unix())})
	return 0, nil
}
