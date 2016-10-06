diskutil unmount daylight
$GOPATH/bin/go-bindata -o="packages/static/static.go" -pkg="static" static/...
GOARCH=amd64  CGO_ENABLED=1  go build -o make_dmg/daylight.app/Contents/MacOS/daylightbin
cd make_dmg
zip -r daylight_osx64.zip daylight.app/Contents/MacOS/daylightbin
./make_dmg.sh -b background.png -i logo-big.icns -s "480:540" -c 240:400:240:200 -n daylight_osx64 "daylight.app"
cd ../
git stash