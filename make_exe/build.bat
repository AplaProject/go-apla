git config --global user.name "Your Name"
git config --global user.email "you@example.com"
go get -u github.com/jteeuwen/go-bindata/...
cd %GOPATH%/src/github.com/EGaaS/go-mvp


rm packages/static/static.go
git stash

go get -u github.com/EGaaS/go-mvp
go install -ldflags "-H windowsgui" github.com/EGaaS/go-mvp
mv C:\go-projects\bin\daylight-go.exe C:\setup\daylight_no_static.exe

rm packages/static/static.go
git stash

go-bindata -o="packages/static/static.go" -pkg="static" static/... 
go install -ldflags "-H windowsgui" github.com/EGaaS/go-mvp
mv C:\go-projects\bin\daylight-go.exe C:\setup\win64\daylight.exe
pause