#!/bin/sh

# linux
ps -ef | grep p2pnode | grep -v grep | cut -c 9-15 | xargs kill -9
# mac
# ps -ef | grep p2pnode | grep -v grep | cut -c 7-11 | xargs kill -9

basedir=/home/ubuntu/ontology/node/dht
cd $basedir

datadir=$basedir/node
rm -rf $basedir/p2pnode*
cp $datadir/node $basedir/p2pnode
rm -rf nohup.out
rm -rf log/test.log

startHttpPort=30000
startNodePort=40000

for idx in $(seq 1 30)
do

name=p2pnode$idx
httpPort=`expr $startHttpPort + $idx`
nodePort=`expr $startNodePort + $idx`
workspace=$datadir
echo "workspace $workspace, httpport $httpPort, nodeport $nodePort"

#cp $datadir/config.json $workspace/config.json
#cp $datadir/log4go.xml $workspace/log4go.xml

nohup ./p2pnode -config=$workspace/config.json \
-log=$workspace/log4go.xml \
-httpport=$httpPort \
-nodeport=$nodePort > $name.log

echo "$name started!"

sleep 1s
done

ps -ef|grep p2pnode
# tail -f nohup.out