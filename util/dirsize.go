package util

import (
	"io/fs"
	"os"
	"path/filepath"
)

func DirSize(path string) int64 {
	var size int64
	filepath.WalkDir(path, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.IsDir() {
			info, err := d.Info()
			if err == nil {
				size += info.Size()
			}
		}
		return nil
	})
	return size
}

func PathExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}
