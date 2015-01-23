package storage

import (
	p "path"

	"io"
	"os"
)

type Store struct {
	path string
}

type File interface {
	io.ReadCloser
	io.ReaderAt
	Stat() (fi os.FileInfo, err error)
}

func Init(path string) (*Store, error) {
	st := new(Store)
	st.path = path

	_, err := os.Stat(path)
	if err != nil {
		err = mkstore(st.path)
	}
	return st, err
}

func (st *Store) Create(id string, name string) (io.WriteCloser, error) {
	path := idPath(st.path, id)
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		return nil, err
	}

	return os.Create(p.Join(path, name))
}

func (st *Store) Store(id string, file io.Reader, name string) (size int64, err error) {
	dest, err := st.Create(id, name)
	if err != nil {
		return 0, err
	}
	defer dest.Close()

	return io.Copy(dest, file)
}

func (st *Store) Get(id string, name string) (File, error) {
	path := idPath(st.path, id)
	return os.Open(p.Join(path, name))
}

func (st *Store) Delete(id string) error {
	path := idPath(st.path, id)
	return os.RemoveAll(path)
}

func (st *Store) del() {
	os.RemoveAll(st.path)
}
