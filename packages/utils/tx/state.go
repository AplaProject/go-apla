package tx

type FileHeader struct {
	Hash     string
	MimeType string
}

type FileField struct {
	FileHeader
	Path string
}

type File struct {
	FileHeader
	Data []byte
}
