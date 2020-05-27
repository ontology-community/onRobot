#!/bin/sh

# linux
# ps -ef | grep p2pnode | grep -v grep | cut -c 9-15 | xargs kill -9
# mac
ps -ef | grep p2pnode | grep -v grep | cut -c 7-11 | xargs kill -9

basedir=/Users/dylen/workspace/gohome/src/github.com/ontology-community/onRobot
cd $basedir
make build-node

datadir=$basedir/target/node
cd $datadir

rm -rf log
mkdir log
rm -rf $basedir/p2pnode*
mv node p2pnode

startHttpPort=10000
startNodePort=20000

for idx in $(seq 1 20)
do

name=p2pnode$idx
httpPort=`expr $startHttpPort + $idx`
nodePort=`expr $startNodePort + $idx`
workspace=$datadir/$name
echo "workspace $workspace, httpport $httpPort, nodeport $nodePort"

mkdir $workspace
mkdir $workspace/log
cp $datadir/config.json $workspace/config.json
cp $datadir/log4go.xml $workspace/log4go.xml

nohup ./p2pnode -config=$workspace/config.json \
-log=$workspace/log4go.xml \
-httpport=$httpPort \
-nodeport=$nodePort &

echo "$name started!"
done

ps -ef|grep p2pnode
# tail -f nohup.out