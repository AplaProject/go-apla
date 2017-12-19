//go:generate sh -c "mockery -inpkg -name Engine -print > file.tmp && mv file.tmp storage_mock.go"
package storage

import "github.com/AplaProject/go-apla/tools/update_server/model"

type Engine interface {
	GetVersionsList() ([]string, error)
	Get(binary model.Build) (model.Build, error)
	Add(binary model.Build) error
	Delete(binary model.Build) error
}
