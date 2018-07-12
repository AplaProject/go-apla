//go:generate sh -c "mockery -inpkg -name KVStorage -print > file.tmp && mv file.tmp kv_storage_mock.go"
package kv

// TODO Delete, Update, Find, Transactions
type KVStorage interface {
	Insert(key, value string) error
	Get(key string) (string, error)
}
