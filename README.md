# onRobot
ontology p2pserver 测试工具

## 使用方式
下载项目onRobot
```bash
https://github.com/ontology-community/onRobot.git
```
根据环境make得到可执行文件

在config.json文件配置节点列表以及magic等参数。
具体测试时，不同的测试用例在params路径下修改用例参数。
然后运行命令如下:
```bash
./robot -t=handshake
```
也支持批量测试
```bash
./robot -t=handshake,heartbeat
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
