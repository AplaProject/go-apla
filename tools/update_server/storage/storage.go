//go:generate sh -c "mockery -inpkg -name Engine -print > file.tmp && mv file.tmp storage_mock.go"
package storage

type Engine interface {
	GetVersionsList() ([]string, error)
	GetBinary(version string) ([]byte, error)
	AddBinary(binary []byte, version string) error
	DeleteBinary(version string) error
}
