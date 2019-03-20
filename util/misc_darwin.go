package util

// Unzip will decompress a zip archive, moving all files and folders
// within the zip file (parameter 1) to an output directory (parameter 2).
func Unzip(src string, dest string) error {
	if err := ExecuteCommand("unzip", "-o", src, "-d", dest); err != nil {
		return err
	}
	// Removes all extended attributes recursively.
	return ExecuteCommand("xattr", "-rc", dest)
}
