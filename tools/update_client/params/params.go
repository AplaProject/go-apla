package params

type ServerParams struct {
	Server   string
	Login    string
	Password string
}

type KeyParams struct {
	PrivateKeyPath string
	PublicKeyPath  string
}

type BinaryParams struct {
	Path       string
	Version    string
	StartBlock int64
}
