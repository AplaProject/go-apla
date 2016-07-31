rm -rf dcoin-go
git config --global user.name "Your Name"
git config --global user.email "you@example.com"
go get -u github.com/jteeuwen/go-bindata/...
#git clone -b dev https://git@github.com/democratic-coin/dcoin-go.git
cd dcoin-go
rm packages/static/static.go
#git stash
#go get -u github.com/democratic-coin/dcoin-go
$GOPATH/bin/go-bindata -o="packages/static/static.go" -pkg="static" static/...
export CGO_ENABLED=1
export GOARCH=amd64 && go build -o dcoin64
export GOARCH=386 && go build -o dcoin32
zip dcoin_freebsd64.zip dcoin64
zip dcoin_freebsd32.zip dcoin32