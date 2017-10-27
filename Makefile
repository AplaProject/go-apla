.DEFAULT_GOAL := all


go-bindata:
	go get -u github.com/jteeuwen/go-bindata/...

static-files:
	rm -rf $GOPATH/src/github.com/AplaProject/go-apla/packages/static/static.go
	${GOPATH}/bin/go-bindata -o="${GOPATH}/src/github.com/AplaProject/go-apla/packages/static/static.go" -pkg="static" -prefix="${GOPATH}/src/github.com/AplaProject/go-apla/" ${GOPATH}/src/github.com/AplaProject/go-apla/static/...

build:
	go build github.com/AplaProject/go-apla

install:
	go install github.com/AplaProject/go-apla

all:
	make go-bindata
	make static-files
	make install
