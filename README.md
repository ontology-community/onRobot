# onRobot
ontology p2pserver 测试工具

## 使用方式
下载项目onRobot \
https://github.com/ontology-community/onRobot.git

在cmd目录下找到需要运行的服务，比如robot，目录树如下:
```dtd
cmd/robot/
├── config.json
├── main.go
├── params
│   ├── AskFakeBlocks.json
│   ├── AttackTxPool.json
│   ├── Connect.json
│   ├── DDOS.json
│   ├── DoubleSpend.json
│   ├── FakePeerID.json
│   ├── HandshakeTimeout.json
│   ├── HandshakeWrongMsg.json
│   ├── Heartbeat.json
│   ├── HeartbeatInterruptPing.json
│   ├── HeartbeatInterruptPong.json
│   ├── ResetPeerID.json
│   └── Transfer.json
├── transfer_wallet.dat
└── wallet.dat
```
其中，config.json是配置文件，params目录下包含具体测试需要的参数

构建
```bash
make build-robot
```
运行
```bash
make robot t=demo
```
也支持批量测试
```bash
make robot t=transferOnt,doubleSpend
```

## 测试用例
```dtd
fakePeerID                          // 伪造peerID
connect                             // 握手
handshakeTimeout                    // 握手超时测试
handshakeWrongMsg                   // 握手客户端发送错误信息
heartbeat                           // 心跳持续测试
heartbeatInterruptPing              // p2p ping中断测试
heartbeatInterruptPong              // p2p pong中断测试
resetPeerID                         // 重置peerID
ddos                                // ddos 建立大量连接并持续保持心跳
askFakeBlocks                       // 伪造blockHeader请求同步 
attackTxPool                        // 交易池攻击
transferOnt                         // ont转账
doubleSpend                         // 双花攻击
txCount                             // 测试p2p轻节点消息转发数量
neighbor                            // 查询邻结点
subnet                              // subnet子网(共识节点及种子节点)
subnetAddMember                     // subnet模拟动态添加共识节点
subnetDelMember                     // subnet模拟动态减少共识节点
subnetGovIsSeed                     // subnet共识节点同时也是种子节点
subnetReserve                       // subnet共识节点添加reserve
```

## 测试条件及结果预期
#### 1、fakePeerID
```dtd
条件:
1.伪造peerID,
2.随机生成pubkey
3.组合pubkey、peerID为PeerKeyID，尝试连接
参数:
{
  "Remote": "172.168.3.158:20338",  // 测试节点
  "DispatchTime": 18                // 持续时间
}
结果:
1.正常连接
解释:
1.handshake过程中，在updatePeerKid时会对peerKeyID进行校验，根据pubkey重新生成peerID
```

#### 2、connect
```dtd
条件:
1.正常生成peerKeyID，握手或在握手时停止于某个步骤
参数:
{
  "Remote": "172.168.3.158:20338",
  "TestCase": 0
}
TestCase:
HandshakeNormal = 0                 // 正常握手
StopClientAfterSendVersion = 1      // 握手时客户端发送version后停止
StopClientAfterReceiveVersion = 2   // 握手时客户端接收version后停止
StopClientAfterUpdateKad = 3        // 握手时客户端更新kad后停止
StopClientAfterReadKad = 4          // 握手时客户端读取kad后停止
StopClientAfterSendAck = 5          // 握手时客户端发送ack后停止
StopServerAfterSendVersion = 6      // 握手时服务端发送version后停止
StopServerAfterReceiveVersion = 7   // 握手时服务端结束到version后停止
StopServerAfterUpdateKad = 8        // 握手时服务端更新kad后停止
StopServerAfterReadKad = 9          // 握手时服务端读取kad后停止
StopServerAfterReadAck = 10         // 握手时服务端接收ack后停止
结果:
a、正常握手连接应该成功
b、握手中断连接应该失败
```

#### 3、handshakeTimeout
```dtd
条件:
a、握手时在某个步骤延时
参数:
{
  "Remote": "172.168.3.158:20338",
  "ClientBlockTime": 20,            // 客户端建立连接时阻塞时间
  "ServerBlockTime": 20,            // 服务端建立连接时阻塞时间
  "Retry": 10                       // 重试次数
}
结果:
a、第一次握手失败
b、第二次握手成功
```

#### 4、handshakeWrongMsg
```dtd
条件:
a、使用参数构造虚假version，并发送到某个目标节点
参数:
{
  "Remote": "172.168.3.158:20338",  // 节点地址  
  "DispatchTime": 20,               // 持续时间
  "Version": 12,                    // version数据结构version字段
  "Services": 36,                   // services字段
  "TimeStamp": 1222123,             // timestamp字段
  "SyncPort": 20338,                // syncPort字段
  "HttpInfoPort": 12,               // httpInfoPort字段
  "Nonce": 128,                     // nonce字段
  "StartHeight":100,                // startHeight字段
  "Relay":1,                        // relay字段
  "IsConsensus": false,             // isConsensus字段
  "SoftVersion":"v1.10.0"           // softVersion字段
}
结果:
a、连接失败
```

#### 5、heartbeat
```dtd
条件:
a、保持正常心跳
参数:
{
  "Remote": "172.168.3.158:20338",
  "InitBlockHeight": 4962,           // 本地模拟初始块高
  "DispatchTime": 20                 // 心跳持续时间
}
结果:
a、连接正常，模拟块高持续增加
```

#### 6、heartbeatInterruptPing
```dtd
条件:
a、心跳过程中，主动中断ping，持续n sec
参数:
{
  "Remote": "172.168.3.158:20338",   // 节点地址 
  "InitBlockHeight": 4962,           // 本地模拟初始块高
  "InterruptAfterStartTime": 20,     // 连接建立后，开始停止发送心跳 
  "InterruptLastTime": 15,           // 心跳停止时间
  "DispatchTime": 60                 // 测试持续时间
}
结果:
a、连接正常，块高保持一定高度后持续增长
解释:
单方面停止ping不会阻断连接
```

#### 7、heartbeatInterruptPong
```dtd
条件:
a、心跳过程中，主动中断pong，持续n sec
参数:
{
  "Remote": "172.168.3.158:20338",
  "InitBlockHeight": 4962,
  "InterruptAfterStartTime": 20,
  "InterruptLastTime": 50,
  "DispatchTime": 120
}
结果:
a、连接正常，块高保持一定高度后持续增长
解释:
单方面停止pong不会阻断连接
```

#### 8、resetPeerID
```dtd
条件:
a、建立连接保持心跳后，变更peerID重连
参数:
{
  "Remote": "172.168.3.158:20338",
  "InitBlockHeight": 4962,
  "DispatchTime": 60
}
结果:
a、连接断开
解释:
connect_controller在beforeHandshakeCheck时会检查连接目的地址，如已存在则抛错
```

#### 9、ddos
```dtd
条件:
a、构造多个虚假peerID
b、与单个目标sync节点距离较近
c、设置节点maxInbound以及maxInBoundPerIP参数
d、虚假peer主动发起连接，并持续ping
参数:
{
  "Remote": "172.168.3.158:20338",
  "JsonRpc": "http://172.168.3.158:20336",
  "InitBlockHeight": 8579,
  "DispatchTime": 120,
  "StartPort": 8000,
  "ConnNumber":128
}
结果:
a、节点正常出块
b、节点dht原邻结点151~165一直存在，重启后也不会被挤出
c、邻结点列表存在大量虚假连接
解释:
连接建立时会先通过connect_controller的逻辑判断，而不是直接进入dht，
当连接数达到maxInBound时，会拒绝后续的连接，而不是替换老的连接.
重启时，bootstrap&recent_peers会并发加载相关节点，
recent_peers内的节点列表头部包含bootstrap内的相关节点。
```

#### 10.askFakeBlocks
```dtd
条件:
a.模拟headerReq请求数据
参数:
{
  "Remote": "172.168.3.162:20338",
  "InitBlockHeight":11000,
  "DispatchTime": 20,
  "StartHash": "d9561c3cfabb06b2df6702c3e278501e9d5545db252fdd40992b4da25ca99a91",   // 模拟block起始hash
  "EndHash": "d9561c3cfabb06b2df6702c3e278501e9d5545db252fdd40992b4da25ca99a90"      // 模拟block结束hash
}
结果:
a.拿不到任何结果
解释:
节点接收到headerReq或者类似请求时会对hash进行校验
```

#### 11.attackTxPool
```dtd
条件:
a、多个恶意节点持续对多个目标seed节点发送大量不合法交易(比如余额不足)
参数:
{
  "RemoteList": [                              // 节点p2p列表
    "172.168.3.158:20338",
    "172.168.3.159:20338",
    "172.168.3.160:20338",
    "172.168.3.161:20338"
  ],
  "JsonRpcList": [                             // 节点rpc列表
    "http://172.168.3.158:20336",
    "http://172.168.3.159:20336",
    "http://172.168.3.160:20336",
    "http://172.168.3.161:20336"
  ],
  "DispatchTime": 10,
  "DestAccount": "AG4pZwKa9cr8ca7PED7FqzUfcwnrQ2N26w", // 转账交易目标账户
  "TxNum": 100141,                                     // 发送的不合法交易数量, txnpool最大容量为10040
  "MinExpectedBlkHeightDiff": 2                        // 测试时间内预期块高度差
}
结果:
a、出块正常
b、测试前后查询余额，账户余额不变
```

#### 12.transferOnt
```dtd
条件:
a.ont转账，为doubleSpend账户准备固定金额，该测试用例一般与doubleSpend组合使用，也可以单独使用
参数:
{
  "Remote": "172.168.3.158:20338",
  "JsonRpc": "http://172.168.3.158:20336",
  "DispatchTime": 5,
  "DestAccount": "AWoQ8oFXXz9EwGBTP2mncqe5ngr1VnKagZ",
  "Amount": 3                                          // 转账额度
}
```

#### 13.doubleSpend
```dtd
条件:
a、单个恶意节点，对多个目标seed节点发送连续的4笔交易，其中1笔能成功，另外3笔不能成功，
   比如只有2块钱的情况下，转账4次，1.1， 1.2， 1.3，1.4
参数:
{
  "RemoteList": [
    "172.168.3.158:20338",
    "172.168.3.159:20338",
    "172.168.3.160:20338",
    "172.168.3.161:20338"
  ],
  "JsonRpcList": [
    "http://172.168.3.158:20336",
    "http://172.168.3.159:20336",
    "http://172.168.3.160:20336",
    "http://172.168.3.161:20336"
  ],
  "DispatchTime": 6,
  "DestAccount": "AG4pZwKa9cr8ca7PED7FqzUfcwnrQ2N26w"
}   
结果:
a、目标seed节点交易池能实时查到这几笔交易
b、测试前后查询余额账户，只转出一笔
```

#### 14.txCount
```dtd
条件:
a、具体网络节点的使用参见 doc/p2pnode.md
b、txCount测试用例配置参数
{
  "IpList": [
    "172.168.3.151",
    ......
    "172.168.3.152",
  ],
  "StartHttpPort": 30001,
  "EndHttpPort": 30003,
  "Remote": "172.168.3.151:40001",
  "DestAccount": "AWoQ8oFXXz9EwGBTP2mncqe5ngr1VnKagZ",
  "SendTicker": 1,
  "StatTicker": 10,
  "TxPerSec": 2,
  "TxPerStat": 10,
  "MsgNumber": 30,
  "StatAfterDuration": 20,
  "Mysql": {
    "Ip": "172.168.3.219",
    "Port": 3306,
    "User": "root",
    "Pwd": "123456",
    "Db": "txstat"
  }
}
因为需要用到多台机器，多个端口构造尽可能多的轻节点，这里我们提供了一个ip列表，
StartHttpPort到EndHttpPort都对应某个ip下的轻节点统计服务。
remote是某个节点的p2p地址，robot通过往这个节点发送消息，实现消息在整个网络的流转。
DestAccount用于构造一笔交易(统计tx时，测试用例构造并发送Tx，该tx为一笔无法完成的转账)，
SendTicker 消息发送间隔
StatTicker 消息统计间隔，统计数据存储到数据库
TxPerSec   每次发送消息数量
TxPerStat  每次统计消息数量
MsgNumber  发送的消息总量
StatAfterDuration 统计滞后于消息发送
Mysql      数据库配置
结果:
以6个节点，持续10s为例
[2020/05/26 10:44:50 CST] [INFO] send tx number 5, recv tx number 5
[2020/05/26 10:44:50 CST] [INFO] send tx number 5, recv tx number 5
[2020/05/26 10:44:50 CST] [INFO] send tx number 5, recv tx number 5
[2020/05/26 10:44:50 CST] [INFO] send tx number 5, recv tx number 5
[2020/05/26 10:44:50 CST] [INFO] send tx number 5, recv tx number 5
[2020/05/26 10:44:50 CST] [INFO] send tx number 6, recv tx number 5
[2020/05/26 10:44:50 CST] [INFO] send tx number 6, recv tx number 5
[2020/05/26 10:44:50 CST] [INFO] send tx number 6, recv tx number 5
[2020/05/26 10:44:50 CST] [INFO] send tx number 6, recv tx number 5
[2020/05/26 10:44:50 CST] [INFO] average send tx number 5.444444, average recv tx number 5.000000, total send tx number 54, total recv tx number 50
[2020/05/26 10:44:50 CST] [DEBG] clear msg stat
[2020/05/26 10:44:50 CST] [INFO] Run Method:txCount success.
[2020/05/26 10:44:50 CST] [INFO] ---------------------------------------------------------------
[2020/05/26 10:44:50 CST] [INFO] 
[2020/05/26 10:44:50 CST] [DEBG] [GC] end testing, stop server and clear instance...
[2020/05/26 10:44:50 CST] [INFO] remove peer 5f0b92825c9b9b23b90e298432c661aaa7bcdd03 from net server
[2020/05/26 10:44:50 CST] [INFO] closing connection: peer 5f0b92825c9b9b23b90e298432c661aaa7bcdd03, address: 127.0.0.1:50394
[2020/05/26 10:44:50 CST] [DEBG] peer disconnected, address: 127.0.0.1:50394, id 14843127213869029055
[2020/05/26 10:44:50 CST] [DEBG] ......
[2020/05/26 10:44:50 CST] [INFO] ===============================================================
[2020/05/26 10:44:50 CST] [INFO] Ontology Tool Finish Total:1 Success:1 Failed:0 Skip:0, SpendTime:13 sec
[2020/05/26 10:44:50 CST] [INFO] ---------------------------------------------------------------
[2020/05/26 10:44:50 CST] [INFO] Success list:
[2020/05/26 10:44:50 CST] [INFO] 1.	txCount
[2020/05/26 10:44:50 CST] [INFO] ===============================================================
整个网络总共发送了54次tx，接收了50次，平均每秒发送5.444444笔交易，接收5笔交易
```

#### 15.neighbor
```dtd
条件:
通过coredns设置reserve1.ontsnip.com为172.168.3.158, reserve2.ontsnip.com为172.168.3.162
在172.168.3.165机器上配置dns服务器，启动服务后，查询邻结点.
通过p2p网络发送FindNodeReq,查询邻结表，直到找到符合预期的节点ip
coredns的使用参见doc/coredns.md
参数:
{
  "Remote": "172.168.3.165:20338",
  "ExpectedIpList": [
    "172.168.3.163"
  ],
  "Timeout": 10
}
结果:
可以找到reserve列表内的节点
```

#### 16.subnet
```dtd
条件:
配置S, G, N(seed,gov,norm数量)为(S, G, N) = (4, 4, 2)的网络拓扑结构，
根据情况设置
所有节点启动并运行到稳态。
参数:
{
  "Subnet":{
    "Seeds":[
      "127.0.0.10:20336",       // mock实现单机多ip地址
      "127.0.0.11:20336",
      "127.0.0.12:20336",
      "127.0.0.13:20336"
    ],
    "Govs":[
      "127.0.0.20:20336",
      "127.0.0.21:20336",
      "127.0.0.22:20336",
      "127.0.0.23:20336"
    ],
    "Norms":[
      "127.0.0.30:20336",
      "127.0.0.31:20336"
    ]
  },
  "SubnetMaxInactiveTime": 600, // 定时清除不活跃的subnet member
  "SubnetRefreshDuration": 1,   // 定时清除并断开非共识节点
  "Dispatch": 15                // 等待到稳态
}
结果:
共识节点subnet member包含自己及其他共识节点，邻结点包含除自己以外的共识节点以及所有种子节点
种子节点subnet member包含所有共识节点，邻接表包含所有共识节点以及除自己以外的所有种子节点
普通节点subnet member为空，邻接表包含除自己以外的其他普通节点以及所有共识节点
```

#### 17.subnetAddMember
```dtd
条件: 
配置(S, G, N) = (4, 4, 2)的网络拓扑结构，并准备两个新的共识节点.
网络节点启动并运行到稳态，然后批量添加共识节点，等待网络再次进入稳态.
参数:
{
  "Subnet":{
    "Seeds":[
      "127.0.0.10:20336",
      "127.0.0.11:20336",
      "127.0.0.12:20336",
      "127.0.0.13:20336"
    ],
    "Govs":[
      "127.0.0.20:20336",
      "127.0.0.21:20336",
      "127.0.0.22:20336",
      "127.0.0.23:20336"
    ],
    "Norms":[
      "127.0.0.30:20336",
      "127.0.0.31:20336"
    ]
  },
  "AddList": [
    "127.0.0.41:20336",             // 待添加的共识节点
    "127.0.0.42:20336"              
  ],
  "SubnetMaxInactiveTime": 600,
  "SubnetRefreshDuration": 1,
  "DispatchBeforeAddGovNode": 15,   // 第一次进入稳态前等待时间
  "DispatchAfterAddGovNode": 90     // 第二次进入稳态前等待时间
}
结果:
同subnet
```

#### 18.subnetDelMember
```dtd
条件: 
配置(S, G, N) = (4, 4, 2)的网络拓扑结构，并准备两个新的共识节点.
网络节点启动并运行到稳态，然后批量添加共识节点，等待网络再次进入稳态.
需要注意的是，delList内的两个共识节点从subnet中删除后，
本身并没有关停服务，而是变成了普通节点，这时候网络结构是(S, G, N) = (4, 2, 4)。
参数:
{
  "Subnet":{
    "Seeds":[
      "127.0.0.10:20336",
      "127.0.0.11:20336",
      "127.0.0.12:20336",
      "127.0.0.13:20336"
    ],
    "Govs":[
      "127.0.0.20:20336",
      "127.0.0.21:20336",
      "127.0.0.22:20336",
      "127.0.0.23:20336"
    ],
    "Norms":[
      "127.0.0.30:20336",
      "127.0.0.31:20336"
    ]
  },
  "DelList": [
    "127.0.0.20:20336",             // 待删除的共识节点，必须存在于govs
    "127.0.0.21:20336"
  ],
  "SubnetMaxInactiveTime": 1,
  "SubnetRefreshDuration": 1,
  "DispatchBeforeDelGovNode": 10,  // 第一次到达稳态前等待时间 
  "DispatchAfterDelGovNode": 15    // 第二次到达稳态前等待时间
}

结果:
同subnet
```

#### 19.subnetGovIsSeed
```dtd
条件: 
配置(S, G, N) = (4, 4, 2)的网络拓扑结构, 其中一个共识节点同时也是种子节点.
网络节点启动并运行到稳态.
参数:
{
  "Subnet":{
    "Seeds":[
      "127.0.0.10:20336",
      "127.0.0.11:20336",
      "127.0.0.12:20336",
      "127.0.0.13:20336"
    ],
    "Govs":[
      "127.0.0.20:20336",
      "127.0.0.21:20336",
      "127.0.0.22:20336",
      "127.0.0.23:20336"
    ],
    "Norms":[
      "127.0.0.30:20336",
      "127.0.0.31:20336"
    ]
  },
  "GovInSeed": "127.0.0.20:20336",   // 既是共识节点又是种子节点，须同时出现在seeds,govs
  "SubnetMaxInactiveTime": 600,
  "SubnetRefreshDuration": 1,
  "Dispatch": 15
}

结果:
该节点邻接表里包含所有种子节点，所有其他共识节点以及所有普通节点。普通节点的subnet member为空。
需要注意的是，程序会先判断其是否为共识节点。
```

#### 20.subnetReserve
```dtd
条件: 
配置(S, G, N) = (4, 4, 2)的网络拓扑结构，
因为seed节点添加reserve没有意义，这里我们尝试在gov及norm节点添加reserve节点列表.
配置其中一个共识节点的rsv为某个普通节点，配置其中一个普通节点的rsv为某个共识节点.
启动并等待网络运行到稳态
参数:
{
  "Subnet":{
    "Seeds":[
      "127.0.0.10:20336",
      "127.0.0.11:20336",
      "127.0.0.12:20336",
      "127.0.0.13:20336"
    ],
    "Govs":[
      "127.0.0.20:20336",
      "127.0.0.21:20336",
      "127.0.0.22:20336",
      "127.0.0.23:20336"
    ],
    "Norms":[
      "127.0.0.30:20336",
      "127.0.0.31:20336"
    ]
  },
  "GovRsv": {                       // 共识节点rsv列表
    "Host": "127.0.0.20",
    "Rsv": ["127.0.0.30"]
  },
  "NormRsv": {                      // 普通节点rsv列表
    "Host": "127.0.0.31",
    "Rsv": ["127.0.0.21"]
  },
  "GovInSeed": "127.0.0.20:20336",
  "SubnetMaxInactiveTime": 600,
  "SubnetRefreshDuration": 1,
  "Dispatch": 90
}

结果:
该测试变化情况较多，这里我们仅通过观察的方式来判断是否正确:
gov节点reserve添加norm节点会使得gov及norm节点的neighborList出现对方
norm节点中添加gov节点，而对方不添加gov，会导致该节点连不上任何节点
```
