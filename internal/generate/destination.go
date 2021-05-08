package generate

import (
	"fmt"
	"os"
	"path/filepath"
)

// MakeFolder from list of dir names
func MakeFolder(names ...string) error {
	return os.MkdirAll(filepath.Join(names...), os.ModePerm)
}

// ValidateFolder ensures a folder exists and can be written
func ValidateFolder(name string) error {
	info, err := os.Stat(name)
	if os.IsNotExist(err) {
		return fmt.Errorf("destination %s not found", name)
	}
	if !info.IsDir() {
		return fmt.Errorf("destination %s is not a folder", name)
	}
	if info.Mode().Perm()&(1<<(uint(7))) == 0 {
		return fmt.Errorf("destination %s not writable", name)
	}
	return nil
}
