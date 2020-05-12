GOFMT=gofmt
GC=go build

VERSION := $(shell git describe --always --tags --long)
BUILD_NODE_PAR = -ldflags "-X main.Version=1.0.0"

ARCH=$(shell uname -m)
SRC_FILES = $(shell git ls-files | grep -e .go$ | grep -v _test.go)

robot: $(SRC_FILES)
	$(GC)  $(BUILD_NODE_PAR) -o robot main.go

robot-cross: robot-windows robot-linux robot-darwin

robot-windows:
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GC) $(BUILD_NODE_PAR) -o robot-windows.exe main.go

robot-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GC) $(BUILD_NODE_PAR) -o robot-linux main.go

robot-darwin:
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GC) $(BUILD_NODE_PAR) -o robot-darwin main.go

tools-cross: tools-windows tools-linux tools-darwin

format:
	$(GOFMT) -w main.go

clean:
	rm -rf *.8 *.o *.out *.6 *exe
	rm -rf robot robot-*

remake:
	make clean && make robot
