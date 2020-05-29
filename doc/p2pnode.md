# p2pnode

## description
p2p轻节点，用于观察验证消息在网络中流转状况，包含以下模块:
* p2p网络, 关闭block sync
* mockTxPool 用于模拟交易池
* stat 用于统计消息收发数量
* httpinfo 用于查询统计数据

## config
在cmd目录下找到需要运行的服务，比如robot，目录树如下:
```dtd
tree cmd/p2pnode/
cmd/p2pnode/
├── config.json
└── main.go

0 directories, 2 files
```
其中，config.json是配置文件, 参考ontology config.json

#### build
```bash
make build-node
```
#### run
```bash
workspace=yourworkspace
httpPort=yourHttpPort
nodePort=yourNodePort

./p2pnode -config=${workspace}/config.json \
-log=${workspace}/log4go.xml \
-httpport=${httpPort} \
-nodeport=${nodePort}
```

## scripts
```dtd
tree scripts
scripts/
├── exec_remote.sh
├── multi_node.sh
└── scp.sh

0 directories, 3 files
```
* scp.sh 上传文件到测试节点集群，可自行修改集群ip列表
* multi_node.sh 在单台机器上运行多个节点，在脚本中修改下列参数, 
num为节点数，workspace为工作目录，startHttpPort为http起始端口，startNodePort为p2p起始端口
```bash
num=8
workspace=/home/ubuntu/ontology/node/dht
startHttpPort=30000
startNodePort=40000
```
* exec_remote.sh 用于运行远程服务器上进程，输入不同方法名实现相关功能:
1. run 运行远端multi_node.sh脚本
2. kill 停止所有p2p节点
3. count 统计p2p节点数量
4. single 当单个节点停止时，启动单个节点
通过修改脚本ipList确定集群，修改single中使用到的ip，idx等参数启动单个节点
