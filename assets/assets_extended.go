package assets

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

func (fs *FileSystem) GetSubDirs(name string) ([]string, error) {
	if !strings.HasSuffix(name, string(filepath.Separator)) {
		name = name + string(filepath.Separator)
	}
	if filepath.Separator != '/' && strings.IndexRune(name, filepath.Separator) >= 0 ||
		strings.Contains(name, "\x00") {
		return nil, errors.New("http: invalid character in file path")
	}
	matchPrefix := strings.TrimRight(name, string(filepath.Separator))

	_, ok := fs.files[name]
	if !ok {
		fileNames := []string{}
		for path, file := range fs.files {
			if filepath.Dir(path) == matchPrefix {
				fi := file.fi
				if fi.IsDir() {
					fileNames = append(fileNames, fi.Name())
				}
			}
		}

		return fileNames, nil
	}

	return nil, os.ErrNotExist
}

func (fs *FileSystem) GetSubFiles(name string) ([]string, error) {
	if !strings.HasSuffix(name, string(filepath.Separator)) {
		name = name + string(filepath.Separator)
	}
	if filepath.Separator != '/' && strings.IndexRune(name, filepath.Separator) >= 0 ||
		strings.Contains(name, "\x00") {
		return nil, errors.New("http: invalid character in file path")
	}
	matchPrefix := strings.TrimRight(name, string(filepath.Separator))

	_, ok := fs.files[name]
	if !ok {
		fileNames := []string{}
		for path, file := range fs.files {
			if filepath.Dir(path) == matchPrefix {
				fi := file.fi
				if !fi.IsDir() {
					fileNames = append(fileNames, fi.Name())
				}
			}
		}

		return fileNames, nil
	}

	return nil, os.ErrNotExist
}
