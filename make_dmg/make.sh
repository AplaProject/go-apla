diskutil unmount Dcoin
$GOPATH/bin/go-bindata -o="packages/static/static.go" -pkg="static" static/...
GOARCH=amd64  CGO_ENABLED=1  go build -o make_dmg/Dcoin.app/Contents/MacOS/dcoinbin
cd make_dmg
zip -r dcoin_osx64.zip Dcoin.app/Contents/MacOS/dcoinbin
./make_dmg.sh -b background.png -i logo-big.icns -s "480:540" -c 240:400:240:200 -n dcoin_osx64 "Dcoin.app"
cd ../
git stash