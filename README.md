### Installation v1.x - only egs-wallet

Install golang >=1.6 https://golang.org/dl/<br>
Set GOPATH<br>
Install git https://git-scm.com/
```
go get -u github.com/jteeuwen/go-bindata/...
git clone -b 1.0 https://github.com/EGaaS/go-egaas-mvp.git
cd go-egaas-mvp
rm -rf packages/static/static.go
$GOPATH/bin/go-bindata -o="packages/static/static.go" -pkg="static" static/..
go build
./go-egaas-mvp
```

### Installation v0.x - full egaas (private blockchain)


Install golang >=1.6 https://golang.org/dl/<br>
Set GOPATH<br>
Install git https://git-scm.com/
```
go get -u github.com/jteeuwen/go-bindata/...
git clone https://github.com/EGaaS/go-egaas-mvp.git
cd go-egaas-mvp
$GOPATH/bin/go-bindata -o="packages/static/static.go" -pkg="static" static/..
go build
./go-egaas-mvp
```


### Questions?
email: hello@egaas.org
