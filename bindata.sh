#!/bin/bash
if [ $# -gt 0 ] && [ $1 = "debug" ] 
then
  DEBUG="-debug=true"
fi
$GOPATH/bin/go-bindata -o="packages/static/static.go" -pkg="static" $DEBUG static/...
if [ -n $DEBUG ] ; then
  find packages/static/ -type f -name "static.go" -exec sed -e 's/\/dc\/static/static/' -i {} \;
fi