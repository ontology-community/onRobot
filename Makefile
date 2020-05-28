# Go parameters
GOCMD=GO111MODULE=on go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test

build-robot:
	rm -rf target/robot
	mkdir target/robot target/robot/params
	cp log4go.xml target/robot/log4go.xml
	cp cmd/robot/config.json target/robot/config.json
	cp cmd/robot/wallet.dat target/robot/wallet.dat
	cp cmd/robot/transfer_wallet.dat target/robot/transfer_wallet.dat
	cp -r cmd/robot/params/* target/robot/params/
	$(GOBUILD) -o target/robot/robot cmd/robot/main.go

build-node:
	rm -rf target/node
	mkdir target/node
	cp log4go.xml target/node/log4go.xml
	cp cmd/p2pnode/config.json target/node/config.json
	$(GOBUILD) -o target/node/node cmd/p2pnode/main.go

build-linux-node:
	GOOS=linux GOARCH=amd64 $(GOBUILD) -o target/node/linux-node cmd/p2pnode/main.go

robot:
	@echo test case $(t)
	./target/robot/robot -config=target/robot/config.json \
	-log=target/robot/log4go.xml \
	-params=target/robot/params \
	-wallet=target/robot/wallet.dat \
	-transfer=target/robot/transfer_wallet.dat \
	-t=$(t)

node:
	@echo httpinfoport p2pport $(httpport) $(nodeport)
	./target/node/node -config=target/node/config.json \
	-log=target/node/log4go.xml \
	-httpport=$(httpport) \
	-nodeport=$(nodeport)

clean:
	rm -rf target/