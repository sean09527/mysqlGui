package storage

import (
	"os"
	"path"

	"github.com/vrischmann/userdir"
)

type LocalStorage struct {
	ConfPath string
}

func NewLocalStorage(filename string) *LocalStorage {
	return &LocalStorage{
		ConfPath: path.Join(userdir.GetConfigHome(), "MyGUI", filename),
	}
}

func (l *LocalStorage) Load() ([]byte, error) {
	d, err := os.ReadFile(l.ConfPath)
	if err != nil {
		return nil, err
	}
	return d, nil
}

func (l *LocalStorage) Save(data []byte) error {

	dir := path.Dir(l.ConfPath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			return err
		}
	}

	return os.WriteFile(l.ConfPath, data, 0777)
}
