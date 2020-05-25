#!/bin/bash

#startHttpPort=10030
#startNodePort=20030

basedir=target/multi_nodes
rm -rf $basedir
mkdir $basedir

for idx in $(seq 1 6)
do
workspace=$basedir/node${idx}
mkdir $workspace $workspace/log
cp target/node/* $workspace/

#let httpPort=startHttpPort+$idx
#let nodePort=startNodePort+$idx

#cd $workspace
#echo "$workspace running..."
#./node -config=config.json \
#	-log=log4go.xml \
#	-httpport=$startHttpPort \
#	-nodeport=$startNodePort

done