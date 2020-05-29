#!/bin/bash

#file=./target/node/config.json
#remote=/home/ubuntu/ontology/node/dht/node/config.json

# file=./log4go.xml
# remote=/home/ubuntu/ontology/node/dht/node/log4go.xml

file=./scripts/multi_node.sh
remote=/home/ubuntu/ontology/node/dht/multi_node.sh

#file=./cmd/robot/config.json
#remote=/home/ubuntu/dht/config.json

#file=./target/robot/params/TxCount.json
#remote=/home/ubuntu/dht/params/TxCount.json

#file=./target/robot/log4go.xml
#remote=/home/ubuntu/dht/log4go.xml

#file=./target/robot/transfer_wallet.dat
#remote=/home/ubuntu/dht/transfer_wallet.dat

#ipList="172.168.3.59"

ipList="\
172.168.3.151 \
172.168.3.152 \
172.168.3.153 \
172.168.3.154 \
172.168.3.155 \
172.168.3.156 \
172.168.3.157 \
172.168.3.158 \
172.168.3.159 \
172.168.3.160 \
172.168.3.161 \
172.168.3.162 \
172.168.3.163 \
172.168.3.164 \
172.168.3.165"

for ip in ${ipList}; do
	scp ${file} root@${ip}:${remote}
done
