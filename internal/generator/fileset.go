package generator

import "fmt"

// FileSet set of files for finding recursion
type FileSet struct {
	filemap map[string]bool
	files   []string
	Err     error
}

// NewFileSet constructor
func NewFileSet() *FileSet {
	return &FileSet{
		filemap: make(map[string]bool),
		files:   []string{},
	}
}

// Add a file to the fileset
func (fset *FileSet) Add(file string) *FileSet {
	_, found := fset.filemap[file]
	if found {
		fset.Err = fmt.Errorf("resursive files at %q", file)
		return fset
	}
	fset.filemap[file] = true
	fset.files = append(fset.files, file)
	return fset
}

// ReverseFiles files found in reverse order
func (fset *FileSet) ReverseFiles() []string {
	res := []string{}
	size := len(fset.files)
	for i := size - 1; i > -1; i-- {
		res = append(res, fset.files[i])
	}
	return res
}
