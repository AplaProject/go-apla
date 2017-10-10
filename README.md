[![Go Report Card](https://goreportcard.com/badge/github.com/AplaProject/go-apla/)](https://goreportcard.com/report/github.com/AplaProject/go-apla/)

### Installation

Install golang >=1.6 https://golang.org/dl/<br>
Set GOPATH<br>
Install git https://git-scm.com/
```
go get -u github.com/jteeuwen/go-bindata/...
go get -u github.com/AplaProject/go-apla
cd $GOPATH/src/github.com/AplaProject/go-apla
$GOPATH/bin/go-bindata -o="packages/static/static.go" -pkg="static" static/..
go build
./go-apla
```


### Questions?
email: hello@apla.io
