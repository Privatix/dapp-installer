package util

import (
	"fmt"
	"os/user"
)

// Unzip will decompress a zip archive, moving all files and folders
// within the zip file (parameter 1) to an output directory (parameter 2).
func Unzip(src, dest, uid string) error {
	if err := ExecuteCommand("unzip", "-o", src, "-d", dest); err != nil {
		return err
	}
	// Removes all extended attributes recursively.
	if err := ExecuteCommand("xattr", "-rc", dest); err != nil {
		return err
	}

	if uid == "" {
		return nil
	}

	// If uid is provided, change files permissions and owner.
	u, err := user.LookupId(uid)
	if err != nil {
		return fmt.Errorf("could not find user by uid `%v`: %v", uid, err)
	}

	if err := ExecuteCommand("chown", "-R", u.Username, dest); err != nil {
		return fmt.Errorf("could not change files owner: %v", err)
	}

	if err := ExecuteCommand("chmod", "-R", "755", dest); err != nil {
		return fmt.Errorf("could not change file permissions: %v", err)
	}

	return nil
}
