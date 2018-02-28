package storage

import (
	"io/ioutil"

	"os"

	"path/filepath"

	"fmt"

	"github.com/GenesisKernel/go-genesis/tools/update_server/model"
	"github.com/pkg/errors"
)

type BinaryStorage struct {
	storagePath string
}

func NewBinaryStorage(storagePath string) BinaryStorage {
	return BinaryStorage{
		storagePath: storagePath,
	}
}

func (bs *BinaryStorage) SaveBuild(build model.Build) error {
	if _, err := os.Stat(bs.storagePath); os.IsNotExist(err) {
		err := os.Mkdir(bs.storagePath, 0770)
		if err != nil {
			return errors.Wrapf(err, "creating directory %s", bs.storagePath)
		}
	}

	err := ioutil.WriteFile(
		filepath.Join(bs.storagePath, build.String()),
		build.Body,
		os.FileMode(0666),
	)
	if err != nil {
		return errors.Wrapf(err, "saving binary to file")
	}

	fmt.Println("file saved?")
	return nil
}

func (bs *BinaryStorage) GetBinary(build model.Build) ([]byte, error) {
	b, err := ioutil.ReadFile(filepath.Join(bs.storagePath, build.String()))
	if err != nil {
		return nil, errors.Wrapf(err, "reading binary body from file")
	}

	return b, nil
}

func (bs *BinaryStorage) DeleteBinary(build model.Build) error {
	fpath := filepath.Join(bs.storagePath, build.String())
	err := os.Remove(fpath)
	if err != nil {
		return errors.Wrapf(err, "removing file %s", fpath)
	}
	return nil
}
