package storage

import p "path"

import (
	"os"
)

const (
	dir_depth = 2
	encodeURL = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_"
)

func mkstore(path string) error {
	return _mkstore(path, dir_depth)
}

func _mkstore(path string, depth int) error {
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil || depth == 0 {
		return err
	}

	for _, l := range encodeURL {
		next_path := p.Join(path, string(l))
		err = _mkstore(next_path, depth-1)
		if err != nil {
			return err
		}
	}
	return nil
}

func idPath(storePath string, id string) string {
	path := storePath
	for i := 0; i < dir_depth; i++ {
		dir := string(id[i])
		path = p.Join(path, dir)
	}
	path = p.Join(path, id)
	return path
}
