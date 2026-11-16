package settings

import "testing"

func TestBrandingDisplayName(t *testing.T) {
	for _, legacy := range []string{"", "File Browser", "Filebrowser", "FileBrowser", "File-Browser", "Neko Filebrowser", "Neko-Filebrowser", "  FILE BROWSER  "} {
		if got := (Branding{Name: legacy}).DisplayName(); got != "FileNeko" {
			t.Errorf("legacy brand %q displayed as %q", legacy, got)
		}
	}
	for _, custom := range []string{"FileNeko", "My Magic Library", "我的魔法书库", "File Browser Club"} {
		if got := (Branding{Name: custom}).DisplayName(); got != custom {
			t.Errorf("custom brand %q replaced with %q", custom, got)
		}
	}
}
