rm -rf daylight-go
git config --global user.name "Your Name"
git config --global user.email "you@example.com"
go get -u github.com/jteeuwen/go-bindata/...
#git clone -b dev https://git@github.com/EGaaS/go-mvp.git
cd daylight-go
rm packages/static/static.go
#git stash
#go get -u github.com/EGaaS/go-mvp
$GOPATH/bin/go-bindata -o="packages/static/static.go" -pkg="static" static/...
export CGO_ENABLED=1
export GOARCH=amd64 && go build -o daylight64
export GOARCH=386 && go build -o daylight32
zip daylight_freebsd64.zip daylight64
zip daylight_freebsd32.zip daylight32