package fbhttp

import (
	"encoding/json"
	"net/http"

	"github.com/filebrowser/filebrowser/v2/auth"
	fberrors "github.com/filebrowser/filebrowser/v2/errors"
	"github.com/filebrowser/filebrowser/v2/rules"
	"github.com/filebrowser/filebrowser/v2/settings"
	"github.com/filebrowser/filebrowser/v2/users"
)

type settingsData struct {
	CurrentPassword       string                `json:"current_password,omitempty"`
	Signup                bool                  `json:"signup"`
	HideLoginButton       bool                  `json:"hideLoginButton"`
	CreateUserDir         bool                  `json:"createUserDir"`
	MinimumPasswordLength uint                  `json:"minimumPasswordLength"`
	UserHomeBasePath      string                `json:"userHomeBasePath"`
	Defaults              settings.UserDefaults `json:"defaults"`
	AuthMethod            settings.AuthMethod   `json:"authMethod"`
	Rules                 []rules.Rule          `json:"rules"`
	Branding              settings.Branding     `json:"branding"`
	Tus                   settings.Tus          `json:"tus"`
	Shell                 []string              `json:"shell"`
	Commands              map[string][]string   `json:"commands"`
	Fail2Ban              *settings.Fail2Ban    `json:"fail2ban,omitempty"`
	EnableWebDAV          *bool                 `json:"enableWebDAV,omitempty"`
}

var settingsGetHandler = withAdmin(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	policy := d.settings.PasswordBanPolicy()
	enableWebDAV := d.settings.WebDAVEnabled(d.server)
	data := &settingsData{
		Fail2Ban:              &policy,
		EnableWebDAV:          &enableWebDAV,
		Signup:                d.settings.Signup,
		HideLoginButton:       d.settings.HideLoginButton,
		CreateUserDir:         d.settings.CreateUserDir,
		MinimumPasswordLength: d.settings.MinimumPasswordLength,
		UserHomeBasePath:      d.settings.UserHomeBasePath,
		Defaults:              d.settings.Defaults,
		AuthMethod:            d.settings.AuthMethod,
		Rules:                 d.settings.Rules,
		Branding:              d.settings.Branding,
		Tus:                   d.settings.Tus,
		Shell:                 d.settings.Shell,
		Commands:              d.settings.Commands,
	}

	return renderJSON(w, r, data)
})

var settingsPutHandler = withAdmin(func(_ http.ResponseWriter, r *http.Request, d *data) (int, error) {
	req := &settingsData{}
	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		return http.StatusBadRequest, err
	}

	if d.settings.AuthMethod == auth.MethodJSONAuth && !users.CheckPwd(req.CurrentPassword, d.user.Password) {
		return http.StatusBadRequest, fberrors.ErrCurrentPasswordIncorrect
	}
	if req.Fail2Ban != nil {
		if err := req.Fail2Ban.Validate(); err != nil {
			return http.StatusBadRequest, err
		}
		d.settings.Fail2Ban = req.Fail2Ban
	}
	if req.EnableWebDAV != nil {
		d.settings.EnableWebDAV = req.EnableWebDAV
	}

	d.settings.Signup = req.Signup
	d.settings.CreateUserDir = req.CreateUserDir
	d.settings.MinimumPasswordLength = req.MinimumPasswordLength
	d.settings.UserHomeBasePath = req.UserHomeBasePath
	d.settings.Defaults = req.Defaults
	d.settings.Rules = req.Rules
	d.settings.Branding = req.Branding
	d.settings.Tus = req.Tus
	d.settings.Shell = req.Shell
	d.settings.Commands = req.Commands
	d.settings.HideLoginButton = req.HideLoginButton

	err = d.store.Settings.Save(d.settings)
	return errToStatus(err), err
})
