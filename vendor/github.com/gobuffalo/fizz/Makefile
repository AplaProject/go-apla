TAGS ?= "sqlite"
GO_BIN ?= "go"

install: packr
	$(GO_BIN) install -tags ${TAGS} -v .
	make tidy

tidy:
ifeq ($(GO111MODULE),on)
	$(GO_BIN) mod tidy
else
	echo skipping go mod tidy
endif

deps:
	$(GO_BIN) get -tags ${TAGS} -t ./...
	make tidy

build: packr
	$(GO_BIN) build -v .
	make tidy

test: packr
	./test.sh
	make tidy

ci-deps: packr
	$(GO_BIN) get -tags ${TAGS} -t ./...

ci-test: packr
	$(GO_BIN) test -tags ${TAGS} -race ./...

lint:
	go get github.com/golangci/golangci-lint/cmd/golangci-lint
	golangci-lint run --enable-all
	make tidy

update:
ifeq ($(GO111MODULE),on)
	rm go.*
	$(GO_BIN) mod init
	$(GO_BIN) mod tidy
else
	$(GO_BIN) get -u -tags ${TAGS}
endif
	make test
	make install
	make tidy

release-test: packr
	$(GO_BIN) test -tags ${TAGS} -race ./...
	make tidy

release:
	$(GO_BIN) get github.com/gobuffalo/release
	make tidy
	release -y -f version.go
	make tidy

packr:
	$(GO_BIN) get github.com/gobuffalo/packr/v2/packr2
	packr2
	make tidy



