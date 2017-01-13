### Installation v1.x

Install golang >=1.6 https://golang.org/dl/
Set GOPATH
```
go get -u github.com/jteeuwen/go-bindata/...
git clone -b 1.0 https://github.com/EGaaS/go-egaas-mvp.git
cd go-egaas-mvp
rm -rf packages/static/static.go
$GOPATH/bin/go-bindata -o="packages/static/static.go" -pkg="static" static/..
go build
./go-egaas-mvp
```

### Installation v0.x


Install golang >=1.6 https://golang.org/dl/
Set GOPATH
```
go get -u github.com/jteeuwen/go-bindata/...
git clone https://github.com/EGaaS/go-egaas-mvp.git
cd go-egaas-mvp
$GOPATH/bin/go-bindata -o="packages/static/static.go" -pkg="static" static/..
go build
./go-egaas-mvp
```
