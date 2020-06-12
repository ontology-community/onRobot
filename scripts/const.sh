#!/bin/bash

# set ip list
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

# set test case var
num=2

# set worksapce which contain target/node
workspace=/home/ubuntu/ontology/node/dht
walletpwd=123456

# set http and p2p port config
startHttpPort=30000
startNodePort=40000