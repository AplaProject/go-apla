#! /bin/bash -e
ARCH0=""
ARCH1="386"
if [ $# -gt 0 ] && [ $1 = "amd64" ]
then
  ARCH0="64"
  ARCH1="amd64"
fi

rm -rf dcoin-go
git config --global user.name "Your Name"
git config --global user.email "you@example.com"
go get -u github.com/jteeuwen/go-bindata/...
rm packages/static/static.go
git stash
go get -u -f github.com/democratic-coin/dcoin-go
go-bindata -o="packages/static/static.go" -pkg="static" static/...
GOARCH=$ARCH1  CGO_ENABLED=1  go build -o make_deb/dcoin$ARCH0/usr/share/dcoin/dcoin
cd make_deb
