#! /bin/bash -e
ARCH0=""
ARCH1="386"
if [ $# -gt 0 ] && [ $1 = "amd64" ]
then
  ARCH0="64"
  ARCH1="amd64"
fi

rm -rf daylight-go
git config --global user.name "Your Name"
git config --global user.email "you@example.com"
go get -u github.com/jteeuwen/go-bindata/...
rm packages/static/static.go
git stash
go get -u -f github.com/AplaProject/go-apla
go-bindata -o="packages/static/static.go" -pkg="static" static/...
GOARCH=$ARCH1  CGO_ENABLED=1  go build -o make_deb/daylight$ARCH0/usr/share/daylight/daylight
cd make_deb
