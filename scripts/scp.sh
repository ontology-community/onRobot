#!/bin/bash

. const.sh

file=./multi_node.sh
remote=/home/ubuntu/ontology/node/dht/multi_node.sh

for ip in ${ipList}; do
	scp ${file} root@${ip}:${remote}
done
