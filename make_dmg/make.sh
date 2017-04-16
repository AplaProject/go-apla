diskutil unmount egaas
$GOPATH/bin/go-bindata -o="packages/static/static.go" -pkg="static" static/...
GOARCH=amd64  CGO_ENABLED=1  go build -ldflags -s -o make_dmg/egaas.app/Contents/MacOS/egaasbin
cd make_dmg
zip -r egaas_osx64.zip egaas.app/Contents/MacOS/egaasbin
./make_dmg.sh -b background.png -i logo-big.icns -s "480:540" -c 240:400:240:200 -n egaas_osx64 "egaas.app"
cd ../