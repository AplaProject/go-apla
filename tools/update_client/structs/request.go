package structs

type Request struct {
	Login string
	Pass  string
	B     Binary
}

func (r *Request) CheckLogin(correctLogin string, correctPass string) bool {
	return correctLogin == r.Login && correctPass == r.Pass
}
