package generator

import "fmt"

type FileSet struct {
	filemap map[string]bool
	files   []string
	Err     error
}

func NewFileSet() *FileSet {
	return &FileSet{
		filemap: make(map[string]bool),
		files:   []string{},
	}
}

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

func (fset *FileSet) Range() []string {
	return fset.files
}

func (fset *FileSet) Reverse() []string {
	res := []string{}
	size := len(fset.files)
	for i := size - 1; i > -1; i-- {
		res = append(res, fset.files[i])
	}
	return res
}
