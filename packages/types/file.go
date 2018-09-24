package types

type File map[string]interface{}

func NewFile() File {
	return File{
		"Name":     "",
		"MimeType": "",
		"Body":     []byte{},
	}
}

func NewFileFromMap(m map[interface{}]interface{}) (f File, ok bool) {
	f = NewFile()

	if f["Name"], ok = m["Name"].(string); !ok {
		return
	}
	if f["MimeType"], ok = m["MimeType"].(string); !ok {
		return
	}
	if f["Body"], ok = m["Body"].([]byte); !ok {
		return
	}

	return
}
