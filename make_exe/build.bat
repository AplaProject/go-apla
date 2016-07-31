git config --global user.name "Your Name"
git config --global user.email "you@example.com"
go get -u github.com/jteeuwen/go-bindata/...
cd %GOPATH%/src/github.com/democratic-coin/dcoin-go


rm packages/static/static.go
git stash

go get -u github.com/democratic-coin/dcoin-go 
go install -ldflags "-H windowsgui" github.com/democratic-coin/dcoin-go
mv C:\go-projects\bin\dcoin-go.exe C:\setup\dcoin_no_static.exe

rm packages/static/static.go
git stash

go-bindata -o="packages/static/static.go" -pkg="static" static/... 
go install -ldflags "-H windowsgui" github.com/democratic-coin/dcoin-go
mv C:\go-projects\bin\dcoin-go.exe C:\setup\win64\dcoin.exe
pause