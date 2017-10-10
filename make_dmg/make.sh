diskutil unmount apla
$GOPATH/bin/go-bindata -o="packages/static/static.go" -pkg="static" static/...
GOARCH=amd64  CGO_ENABLED=1  go build -ldflags -s -o make_dmg/apla.app/Contents/MacOS/aplabin
cd make_dmg
zip -r apla_osx64.zip apla.app/Contents/MacOS/aplabin
./make_dmg.sh -b background.png -i logo-big.icns -s "480:540" -c 240:400:240:200 -n apla_osx64 "apla.app"
cd ../
