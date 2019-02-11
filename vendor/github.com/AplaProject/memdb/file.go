package memdb

import (
	"fmt"
	"io"
	"os"

	"github.com/pkg/errors"
	"github.com/tidwall/resp"
)

var ErrOpenFile = errors.New("opening file")

type fileStorage struct {
	file *os.File
}

type command int8

const (
	commandSET command = iota
	commandDEL
)

type fileItem struct {
	item
	command command
}

func openFileStorage(path string) (*fileStorage, error) {
	var err error
	fs := &fileStorage{}

	fs.file, err = os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return nil, err
	}

	return fs, nil
}

type readResult struct {
	item fileItem
	err  error
}

func (fs *fileStorage) read() chan *readResult {
	results := make(chan *readResult)

	go func() {
		rd := resp.NewReader(fs.file)

		for {
			result := readResult{}
			v, _, err := rd.ReadValue()
			if err == io.EOF {
				break
			}
			if err != nil {
				result.err = err
				break
			}

			if v.Type() == resp.Array {
				for i, v := range v.Array() {
					switch i {
					case 0:
						command := v.String()
						if command == "set" {
							result.item.command = commandSET
						} else if command == "del" {
							result.item.command = commandDEL
						}
					case 1:
						result.item.key = dbKey(v.String())
					case 2:
						result.item.value = v.String()
					}
				}
				results <- &result
			}
		}

		close(results)
	}()

	return results
}

func (fs *fileStorage) write(items ...fileItem) error {
	writer := resp.NewWriter(fs.file)

	for _, item := range items {
		row := make([]resp.Value, 0)

		if item.command == commandSET {
			row = append(row, resp.StringValue("set"), resp.StringValue(string(item.key)), resp.StringValue(item.value))
		} else if item.command == commandDEL {
			row = append(row, resp.StringValue("del"), resp.StringValue(string(item.key)))
		} else {
			panic(fmt.Sprintf("unknwon command %d", item.command))
		}

		if err := writer.WriteArray(row); err != nil {
			return err
		}
	}

	return nil
}

func (fs *fileStorage) close() error {
	return fs.file.Close()
}
