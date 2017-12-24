# Apla auto-updates server

### API Endpoints:

#### /api/v1
##### Private

*POST /private/binary - adding build
Data format:
```
{
	"time": "0001-01-01T00:00:00Z",
	"name": "go-apla",
	"body": "AQID",
	"sign": "blah",

	"number": "1.0",
	"os":"linux",
	"arch": "amd64",

	"start_block": 141278,
	"is_critical": true
}
```

*DELETE /private/binary/{os}/{arch}/{version} - deleting binary


##### Public

*GET /{os}/{arch}/last - get info about last version
*GET /{os}/{arch}/versions - get all versions list
*GET /{os}/{arch}/{version} - get info about specific version
*GET /{os}/{arch}/{version}/binary - download binary of the specific version