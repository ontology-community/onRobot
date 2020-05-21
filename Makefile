# Go parameters
GOCMD=GO111MODULE=on go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test

build-robot:
	rm -rf target/
	mkdir target target/params
	cp log4go.xml target/log4go.xml
	cp cmd/robot/config.json target/config.json
	cp cmd/robot/wallet.dat target/wallet.dat
	cp cmd/robot/transfer_wallet.dat target/transfer_wallet.dat
	cp -r cmd/robot/params/* target/params/
	$(GOBUILD) -o target/robot cmd/robot/main.go

build: build-robot

robot:
	@echo test case $(t)
	./target/robot -config=target/config.json \
	-log=target/log4go.xml \
	-params=target/params \
	-wallet=target/wallet.dat \
	-transfer=target/transfer_wallet.dat \
	-t=$(t)

node:
	$(GOBUILD) -o node cmd/node/main.go

clean:
	rm -rf target/