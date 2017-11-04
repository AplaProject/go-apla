[![Go Report Card](https://goreportcard.com/badge/github.com/AplaProject/go-apla)](https://goreportcard.com/report/github.com/AplaProject/go-apla) 
[![Build Status](https://travis-ci.org/AplaProject/go-apla.svg?branch=master)](https://travis-ci.org/AplaProject/go-apla) 
[![Documentation](https://img.shields.io/badge/docs-latest-brightgreen.svg?style=flat)](http://apla.readthedocs.io/en/latest/)
[![API Reference](
https://camo.githubusercontent.com/915b7be44ada53c290eb157634330494ebe3e30a/68747470733a2f2f676f646f632e6f72672f6769746875622e636f6d2f676f6c616e672f6764646f3f7374617475732e737667
)](https://godoc.org/github.com/AplaProject/go-apla)
[![Gitter](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/go-apla?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge)


# Installation

## Requirements

* Go >=1.9
* git

## Build

Clone:
```
git clone https://github.com/AplaProject/go-apla.git $GOPATH/src/github.com/AplaProject/go-apla
```

Build Apla:
```
go get -u github.com/jteeuwen/go-bindata/...
$GOPATH/bin/go-bindata -o="$GOPATH/src/github.com/AplaProject/go-apla/packages/static/static.go" -pkg="static" -prefix="$GOPATH/src/github.com/AplaProject/go-apla/" $GOPATH/src/github.com/AplaProject/go-apla/static/...
go install github.com/AplaProject/go-apla
```

# Running

Create Apla directory and copy binary:
```
mkdir ~/apla
cp $GOPATH/bin/go-apla ~/apla
```

Run apla:
```
~/apla/go-apla
```

To work through GUI you need to install https://github.com/AplaProject/apla-front

----------


### Questions?
email: hello@apla.io
