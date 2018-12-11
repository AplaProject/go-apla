GOCMD=go
GOGET=$(GOCMD) get -u
GOGENERATE=$(GOCMD) generate
GOTEST=$(GOCMD) test

all: build test

build:
	$(GOGENERATE)

test:
	$(GOCMD) test -run ^TestGrammar$$ -v -count 1

deps:
	$(GOGET) golang.org/x/tools/cmd/goyacc
	$(GOGET) github.com/cznic/golex