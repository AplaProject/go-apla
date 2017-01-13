### Installation v0.x

```
git clone -b 1.0 https://github.com/EGaaS/go-egaas-mvp.git
cd go-egaas-mvp
rm -rf packages/static/static.go
$GOPATH/bin/go-bindata -o="packages/static/static.go" -pkg="static" static/..
go build
./go-egaas-mvp
```

### Installation v1.x

```
git clone https://github.com/EGaaS/go-egaas-mvp.git
cd go-egaas-mvp
$GOPATH/bin/go-bindata -o="packages/static/static.go" -pkg="static" static/..
go build
./go-egaas-mvp
```
