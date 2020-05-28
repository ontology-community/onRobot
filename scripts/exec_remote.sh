#!/bin/bash

ipList="172.168.3.151"

echo input funcname
read func

run() {
for ip in $ipList; do
	ssh root@$ip "cd /home/ubuntu/ontology/node/dht; ./multi_node.sh"
done
}

kill() {
for ip in $ipList; do
	ssh root@$ip "killall -9 p2pnode"
done
}

count() {
for ip in $ipList; do
	ssh root@$ip "ps -ef|grep p2pnode|wc -l"
done
}

${func}