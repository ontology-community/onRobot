/*
 * Copyright (C) 2018 The ontology Authors
 * This file is part of The ontology library.
 *
 * The ontology is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The ontology is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The ontology.  If not, see <http://www.gnu.org/licenses/>.
 */

package methods

import (
	"encoding/json"
	log4 "github.com/alecthomas/log4go"
	common2 "github.com/ontio/ontology/common"
	"github.com/ontology-community/onRobot/common"
	"github.com/ontology-community/onRobot/p2pserver/message/types"
	"github.com/ontology-community/onRobot/p2pserver/net/netserver"
	"github.com/ontology-community/onRobot/p2pserver/protocols"
	"github.com/ontology-community/onRobot/utils/timer"
	"sync"
	"time"
)

const MaxNetServerNumber = 128

var (
	tr     = timer.NewTimer(2)
	nsList = make([]*netserver.NetServer, 0, MaxNetServerNumber)
)

func reset() {
	log4.Debug("[GC] end testing, stop server and clear instance...")
	common.Reset()
	for _, ns := range nsList {
		ns.Stop()
	}
	nsList = make([]*netserver.NetServer, 0, MaxNetServerNumber)
}

func Handshake() bool {
	// 1. get params from json file
	var params struct {
		Remote   string
		TestCase uint8
	}
	if err := getParamsFromJsonFile("Handshake.json", &params); err != nil {
		_ = log4.Error("%s", err)
		return false
	}

	// 2. set common params
	common.SetHandshakeStopLevel(params.TestCase)

	// 3. setup p2p.protocols
	protocol := protocols.NewOnlyHeartbeatMsgHandler()
	ns := GenerateNetServerWithProtocol(protocol)

	// 4. connect and handshake
	if err := ns.Connect(params.Remote); err != nil {
		log4.Debug("connecting to %s failed, err: %s", params.Remote, err)
	} else {
		log4.Info("handshake end!")
	}

	return true
}

func HandshakeWrongMsg() bool {

	// 1. get params from json file
	var params struct {
		Remote   string
		WrongMsg bool
	}
	if err := getParamsFromJsonFile("HandshakeWrongMsg.json", &params); err != nil {
		_ = log4.Error("%s", err)
		return false
	}

	protocol := protocols.NewOnlyHeartbeatMsgHandler()
	ns := GenerateNetServerWithProtocol(protocol)

	common.SetHandshakeWrongMsg(params.WrongMsg)
	if err := ns.Connect(params.Remote); err != nil {
		log4.Debug("connecting to %s failed, err: %s", params.Remote, err)
	} else {
		log4.Info("handshakeWrongMsg end!")
	}

	return true
}

func HandshakeTimeout() bool {
	var params struct {
		Remote string
		ClientBlockTime,
		ServerBlockTime int
		Retry int
	}
	if err := getParamsFromJsonFile("HandshakeClientTimeout.json", &params); err != nil {
		_ = log4.Error("%s", err)
		return false
	}

	protocol := protocols.NewOnlyHeartbeatMsgHandler()
	ns := GenerateNetServerWithProtocol(protocol)

	common.SetHandshakeClientTimeout(params.ClientBlockTime)
	common.SetHandshakeServerTimeout(params.ServerBlockTime)
	if err := ns.Connect(params.Remote); err != nil {
		log4.Debug("connecting to %s failed, err: %s", params.Remote, err)
	} else {
		log4.Info("handshake success!")
		return true
	}

	for i := 0; i < params.Retry; i++ {
		log4.Debug("connecting retry cnt %d", i)
		common.SetHandshakeClientTimeout(0)
		if err := ns.Connect(params.Remote); err != nil {
			log4.Debug("connecting to %s failed, err: %s", params.Remote, err)
		} else {
			log4.Info("handshake success!")
			return true
		}
	}

	return true
}

func Heartbeat() bool {
	var params struct {
		Remote          string
		InitBlockHeight uint64
		DispatchTime    int
	}
	if err := getParamsFromJsonFile("Heartbeat.json", &params); err != nil {
		_ = log4.Error("%s", err)
		return false
	}

	protocol := protocols.NewOnlyHeartbeatMsgHandler()
	ns := GenerateNetServerWithProtocol(protocol)

	common.SetHeartbeatTestBlockHeight(params.InitBlockHeight)
	if err := ns.Connect(params.Remote); err != nil {
		_ = log4.Error("connecting to %s failed, err: %s", params.Remote, err)
		return false
	}

	dispatch(params.DispatchTime)

	log4.Info("heartbeat end!")
	return true
}

func HeartbeatInterruptPing() bool {
	var params struct {
		Remote                  string
		InitBlockHeight         uint64
		InterruptAfterStartTime int64
		InterruptLastTime       int64
		DispatchTime            int
	}
	if err := getParamsFromJsonFile("HeartbeatInterruptPing.json", &params); err != nil {
		_ = log4.Error("%s", err)
		return false
	}

	common.SetHeartbeatTestBlockHeight(params.InitBlockHeight)
	common.SetHeartbeatTestInterruptAfterStartTime(params.InterruptAfterStartTime)
	common.SetHeartbeatTestInterruptPingLastTime(params.InterruptLastTime)

	protocol := protocols.NewOnlyHeartbeatMsgHandler()
	ns := GenerateNetServerWithProtocol(protocol)

	if err := ns.Connect(params.Remote); err != nil {
		_ = log4.Error("connecting to %s failed, err: %s", params.Remote, err)
		return false
	}

	dispatch(params.DispatchTime)

	log4.Info("heartbeat end!")
	return true
}

func HeartbeatInterruptPong() bool {
	var params struct {
		Remote                  string
		InitBlockHeight         uint64
		InterruptAfterStartTime int64
		InterruptLastTime       int64
		DispatchTime            int
	}
	if err := getParamsFromJsonFile("HeartbeatInterruptPong.json", &params); err != nil {
		_ = log4.Error("%s", err)
		return false
	}

	common.SetHeartbeatTestBlockHeight(params.InitBlockHeight)
	common.SetHeartbeatTestInterruptAfterStartTime(params.InterruptAfterStartTime)
	common.SetHeartbeatTestInterruptPongLastTime(params.InterruptLastTime)

	protocol := protocols.NewOnlyHeartbeatMsgHandler()
	ns := GenerateNetServerWithProtocol(protocol)

	if err := ns.Connect(params.Remote); err != nil {
		_ = log4.Error("connecting to %s failed, err: %s", params.Remote, err)
		return false
	}

	dispatch(params.DispatchTime)

	log4.Info("heartbeat end!")
	return true
}

// ResetPeerID 更换peerID，并尝试重连
// 第一次连接成功，更换peerID后重连失败，beforeHandshake检查时会因为connect_controller已包含该remote而失败。
// 单如果只是更换peerID，而不重连，纯心跳服务不会受任何影响。
func ResetPeerID() bool {
	var params struct {
		Remote          string
		InitBlockHeight uint64
		DispatchTime    int
	}
	if err := getParamsFromJsonFile("ResetPeerID.json", &params); err != nil {
		_ = log4.Error("%s", err)
		return false
	}

	common.SetHeartbeatTestBlockHeight(params.InitBlockHeight)
	protocol := protocols.NewOnlyHeartbeatMsgHandler()
	ns := GenerateNetServerWithProtocol(protocol)

	if err := ns.Connect(params.Remote); err != nil {
		_ = log4.Error("connecting to %s failed, err: %s", params.Remote, err)
		return false
	}
	oldPeerID := ns.GetID()
	log4.Debug("old peerID %s", oldPeerID.ToHexString())

	if err := ns.ResetRandomPeerID(); err != nil {
		_ = log4.Error("%s", err)
		return false
	}
	newPeerID := ns.GetID()
	log4.Debug("new peerID %s", newPeerID.ToHexString())
	if err := ns.Connect(params.Remote); err != nil {
		_ = log4.Error("connecting to %s failed, err: %s", params.Remote, err)
		return true
	}

	dispatch(params.DispatchTime)

	log4.Info("reset peerID end!")
	return true
}

// ddos 攻击, 构造多个peerID，连接并发送ping
// 结果: 对于同一ip，节点最多接收16个链接，其他的链接会在客户端进行连接时失败，并且不影响块同步速度
func DDos() bool {

	// try to get blockheight
	var params struct {
		Remote          string
		InitBlockHeight uint64
		DispatchTime    int
		StartPort       int
		ConnNumber      int
	}
	if err := getParamsFromJsonFile("DDOS.json", &params); err != nil {
		_ = log4.Error("%s", err)
		return false
	}

	common.Initialize()
	height, err := common.GetBlockHeight()
	if err != nil {
		_ = log4.Error("%s", err)
		return false
	} else {
		log4.Debug("block height before ddos %d", height)
	}

	portlock := new(sync.Mutex)
	common.SetHeartbeatTestBlockHeight(params.InitBlockHeight)
	for i := 0; i < params.ConnNumber; i++ {
		go func(port uint16, mtx *sync.Mutex) {
			protocol := protocols.NewOnlyHeartbeatMsgHandler()
			ns := GenerateNetServerWithContinuePort(protocol, port, mtx)
			peerID := ns.GetID()
			if err := ns.Connect(params.Remote); err != nil {
				_ = log4.Error("peer %s connecting to %s failed, err: %s", peerID.ToHexString(), params.Remote, err)
			} else {
				log4.Debug("peer %s, index %d connecting to %s success", peerID.ToHexString(), int(port)-params.StartPort, params.Remote)
			}
		}(uint16(params.StartPort+i), portlock)
	}

	go func() {
		ticker := time.NewTicker(3 * time.Second)
		for {
			select {
			case <-ticker.C:
				height2, err := common.GetBlockHeight()
				if err != nil {
					log4.Debug("get block height failed,%s", err)
				} else {
					log4.Debug("block height during ddos %d", height2)
				}
			}
		}
	}()

	dispatch(params.DispatchTime)

	log4.Info("ddos attack end!")
	return true
}

// 异常块高
// 无法建立可以通过的blockHash，同步节点接收到后会丢掉异常blockhash
func AskFakeBlocks() bool {
	var params struct {
		Remote             string
		InitBlockHeight    uint64
		DispatchTime       int
		StartHash, EndHash string
	}
	if err := getParamsFromJsonFile("AskFakeBlocks.json", &params); err != nil {
		_ = log4.Error("%s", err)
		return false
	}

	common.SetHeartbeatTestBlockHeight(params.InitBlockHeight)
	protocol := protocols.NewOnlyHeartbeatMsgHandler()
	ns := GenerateNetServerWithProtocol(protocol)
	peer, err := ns.ConnectAndReturnPeer(params.Remote)
	if err != nil {
		_ = log4.Error("connecting to %s failed, err: %s", params.Remote, err)
		return false
	}

	startHash, _ := common2.Uint256FromHexString(params.StartHash)
	endHash, _ := common2.Uint256FromHexString(params.EndHash)
	req := &types.HeadersReq{
		HashStart: startHash,
		HashEnd:   endHash,
		Len:       1,
	}
	if err = peer.Send(req); err != nil {
		_ = log4.Error("send headersReq failed, err %s", err)
		return false
	}

	// dispatch
	if msg := protocol.Out(params.DispatchTime); msg != nil {
		log4.Debug("invalid block endHash accepted by sync node, msg %v", msg)
		return false
	} else {
		log4.Info("invalid block endHash rejected by sync node")
		return true
	}
}

// todo
// 非法交易攻击
func AttackTxPool() bool {
	//type TestJson struct {
	//	localid int
	//	Data string
	//}
	//
	//tj := &TestJson{
	//	localid: 100,
	//	Data: "hello",
	//}
	var tj []byte = nil
	bz, _ := json.Marshal(tj)
	log4.Info("----------- %s", string(bz))
	return true
}

// todo
// 双花
func DoubleSpend() bool {
	return true
}

// todo
// 路由表攻击
func AttackRoutable() bool {
	return true
}
