#!/bin/sh

ps -ef | grep p2pnode | grep -v grep | cut -c 9-15 | xargs kill -9

sleep 5s

basedir=/Users/dylen/workspace/gohome/src/github.com/ontology-community/onRobot/target
datadir=$basedir/node

rm -rf $basedir/p2pnode*

startHttpPort=10030
startNodePort=20030

for idx in $(seq 1 6)
do

name=p2pnode$idx
httpPort=`expr $startHttpPort + $idx`
nodePort=`expr $startNodePort + $idx`
workspace=$basedir/$name
echo "workspace $workspace, httpport $httpPort, nodeport $nodePort"

sudo mkdir $workspace
sudo mkdir $workspace/log
cp $datadir/node $workspace/p2pnode
cp $datadir/config.json $workspace/config.json
cp $datadir/log4go.xml $workspace/log4go.xml

cd $workspace
pwd

nohup ./p2pnode -config=$workspace/config.json \
-log=$workspace/log4go.xml \
-httpport=$httpPort \
-nodeport=$nodePort & > $name.log

echo "$name started!"
sleep 10s
done