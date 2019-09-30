package util

// Unzip will decompress a zip archive, moving all files and folders
// within the zip file (parameter 1) to an output directory (parameter 2).
func Unzip(src, dest, _ string) error {
	return ExecuteCommand("tar", "xpf", src, "-C", dest, "--numeric-owner")
}
