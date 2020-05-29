#!/bin/bash

. const.sh

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
    echo input ip
    read ip
    echo input node index
    read idx

    httpPort=`expr ${startHttpPort} + ${idx}`
    nodePort=`expr ${startNodePort} + ${idx}`
    name=p2pnode${idx}
    cmd="nohup ./p2pnode -config=${workspace}/config.json -log=${workspace}/log4go.xml -httpport=${httpPort} -nodeport=${nodePort} > $name.log &"
    ssh root@${ip} "cd ${workspace}; rm -rf ${name}.log; ${cmd}"
}

${func}