SOURCEDIR=.
SOURCES := $(shell find . -name '*.go')

BINARY=kubenforce

.DEFAULT_GOAL: $(BINARY)

$(BINARY): $(SOURCES)
		CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ${BINARY} main.go
		curl -o ca-certificates.crt https://raw.githubusercontent.com/bagder/ca-bundle/master/ca-bundle.crt

.PHONY: install
install:
		go install ./...

.PHONY: clean
clean:
		if [ -f ${BINARY} ] ; then rm ${BINARY} ; fi
