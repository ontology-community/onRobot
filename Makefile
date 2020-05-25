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

build-node:
	rm -rf target/
	mkdir target
	cp log4go.xml target/log4go.xml
	cp cmd/p2pnode/config.json target/config.json
	$(GOBUILD) -o target/node cmd/p2pnode/main.go

robot:
	@echo test case $(t)
	./target/robot -config=target/config.json \
	-log=target/log4go.xml \
	-params=target/params \
	-wallet=target/wallet.dat \
	-transfer=target/transfer_wallet.dat \
	-t=$(t)

node:
	@echo httpinfoport p2pport $(httpinfoport) $(nodeport)
	./target/node -config=target/config.json \
	-log=target/log4go.xml \
	-httpport=$(httpinfoport) \
	-nodeport=$(nodeport)

clean:
	rm -rf target/