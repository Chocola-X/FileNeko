package settings

import "strings"

const DefaultBrandingName = "FileNeko"

// DisplayName upgrades the built-in brand while preserving administrator names.
func (b Branding) DisplayName() string {
	switch strings.ToLower(strings.TrimSpace(b.Name)) {
	case "", "file browser", "filebrowser", "file-browser", "neko filebrowser", "neko-filebrowser":
		return DefaultBrandingName
	default:
		return b.Name
	}
}

// Branding contains the branding settings of the app.
type Branding struct {
	Name                  string `json:"name"`
	DisableExternal       bool   `json:"disableExternal"`
	DisableUsedPercentage bool   `json:"disableUsedPercentage"`
	Files                 string `json:"files"`
	Theme                 string `json:"theme"`
	Color                 string `json:"color"`
}
