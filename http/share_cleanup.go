package fbhttp

import (
	"errors"
	"path/filepath"
	"strings"

	fberrors "github.com/filebrowser/filebrowser/v2/errors"
	"github.com/filebrowser/filebrowser/v2/files"
)

// revokeSharesForPath matches physical files, not scope-relative names. Merely
// dropping the owner filter would let users revoke unrelated scopes' shares.
// Revoke before mutation and fail closed if cleanup cannot be completed.
func revokeSharesForPath(d *data, name string) error {
	target, err := files.ResolvePath(d.user.FullPath(name))
	if err != nil {
		return err
	}
	links, err := d.store.Share.All()
	if errors.Is(err, fberrors.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	for _, link := range links {
		owner := d.user
		if link.UserID != owner.ID {
			owner, err = d.store.Users.Get(d.server.Root, d.server.FollowExternalSymlinks, link.UserID)
			if errors.Is(err, fberrors.ErrNotExist) {
				continue
			}
			if err != nil {
				return err
			}
		}
		candidate, err := files.ResolvePath(owner.FullPath(link.Path))
		if err != nil {
			return err
		}
		if d.server.CaseInsensitiveFs {
			candidate = strings.ToLower(candidate)
		}
		compare := target
		if d.server.CaseInsensitiveFs {
			compare = strings.ToLower(compare)
		}
		rel, err := filepath.Rel(compare, candidate)
		if err != nil {
			continue
		} // Different volumes cannot refer to the same file.
		if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			continue
		}
		if err := d.store.Share.Delete(link.Hash); err != nil {
			return err
		}
	}
	return nil
}
