#!/bin/sh

killall -9 p2pnode

workspace=/home/ubuntu/ontology/node/dht
cd ${workspace}

rm -rf log
mkdir log
rm -rf ${workspace}/p2pnode*
rm -rf nohup.out
cp ${workspace}/node/config.json config.json
cp ${workspace}/node/log4go.xml log4go.xml
cp ${workspace}/node/node p2pnode

startHttpPort=30000
startNodePort=40000

for idx in $(seq 1 30)
do
name=p2pnode${idx}
httpPort=`expr ${startHttpPort} + ${idx}`
nodePort=`expr ${startNodePort} + ${idx}`
echo "workspace ${workspace}, httpport ${httpPort}, nodeport ${nodePort}"

nohup ./p2pnode -config=${workspace}/config.json \
-log=${workspace}/log4go.xml \
-httpport=${httpPort} \
-nodeport=${nodePort} > $name.log &

echo "$name started!"

sleep 1s
done

ps -ef|grep p2pnode|wc -l