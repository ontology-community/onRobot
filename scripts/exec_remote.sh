#!/bin/bash

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

echo input funcname
read func

run() {
for ip in ${ipList}; do
	ssh root@${ip} "cd /home/ubuntu/ontology/node/dht; ./multi_node.sh"
	echo "${ip} nodes all started!"
done
}

kill() {
for ip in ${ipList}; do
	ssh root@${ip} "killall -9 p2pnode"
done
}

# 统计节点数量
count() {
for ip in ${ipList}; do
    echo "${ip} node number:"
	ssh root@$ip "ps -ef|grep p2pnode|grep -v grep|wc -l"
done
}

# 节点启动不成功时，单独启动
single() {
    ip="172.168.3.151"
    idx=14

    workspace=/home/ubuntu/ontology/node/dht
    startHttpPort=30000
    startNodePort=40000

    httpPort=`expr ${startHttpPort} + ${idx}`
    nodePort=`expr ${startNodePort} + ${idx}`
    name=p2pnode${idx}
    cmd="nohup ./p2pnode -config=${workspace}/config.json -log=${workspace}/log4go.xml -httpport=${httpPort} -nodeport=${nodePort} > $name.log &"
    ssh root@${ip} "cd ${workspace}; rm -rf ${name}.log; ${cmd}"
}

${func}