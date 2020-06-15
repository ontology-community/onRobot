#!/bin/bash

. const.sh

killall -9 p2pnode

# prepare
cd ${workspace}
rm -rf *.log
rm -rf log
rm -rf nohup.out
rm -rf p2pnode
rm -rf wallet.dat
mkdir log

cp ${workspace}/node/config.json config.json
cp ${workspace}/node/node p2pnode
cp ${workspace}/node/wallet.dat wallet.dat

# start nodes
for idx in $(seq 1 ${num})
do
name=p2pnode${idx}
httpPort=`expr ${startHttpPort} + ${idx}`
nodePort=`expr ${startNodePort} + ${idx}`
echo "workspace ${workspace}, httpport ${httpPort}, nodeport ${nodePort}"

nohup ./p2pnode -config=${workspace}/config.json \
-httpport=${httpPort} \
-nodeport=${nodePort} > $name.log &

echo "$name started!"

sleep 1s
done
echo "all started!"

# stat started p2pnodes number
ps -ef|grep p2pnode|grep -v grep|wc -l