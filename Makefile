DLV := $(shell go env GOPATH)/bin/dlv

.PHONY: build debug-build dlv run

build:
	go build -o main ./cmd/

debug-build:
	go build -gcflags="all=-N -l" -o main ./cmd/

dlv: debug-build
	sudo $(DLV) dap --listen=127.0.0.1:2345 --only-same-user=false

run: build
	sudo ./main
