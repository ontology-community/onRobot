#!/bin/bash

basedir=target/multi_nodes
rm -rf $basedir
mkdir $basedir

for idx in $(seq 1 6)
do
workspace=$basedir/node${idx}
mkdir $workspace $workspace/log
cp target/node/* $workspace/

done