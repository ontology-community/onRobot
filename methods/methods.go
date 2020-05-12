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
	log4 "github.com/alecthomas/log4go"
	"github.com/ontology-community/onRobot/common"
	"github.com/ontology-community/onRobot/config"
	"github.com/ontology-community/onRobot/p2pserver/net/netserver"
	"github.com/ontology-community/onRobot/p2pserver/net/protocol"
	"github.com/ontology-community/onRobot/p2pserver/protocols"
	"github.com/ontology-community/onRobot/utils/timer"
)

var (
	ns *netserver.NetServer
	tr *timer.Timer
)

func setup(protocol p2p.Protocol) {
	var err error

	if ns, err = netserver.NewNetServer(protocol, config.DefConfig.Net); err != nil {
		log4.Crashf("[NewNetServer] crashed, err %s", err)
	}
	if err = ns.Start(); err != nil {
		log4.Crashf("start netserver failed, err %s", err)
	}

	tr = timer.NewTimer(2)
}

func reset() {
	log4.Debug("[GC] end testing, stop server and clear instance...")
	ns.Stop()
	common.Reset()
	ns = nil
	tr = nil
}

// methods
func Demo() bool {
	log4.Info("hello, dht demo")
	return true
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
	setup(protocol)

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
	setup(protocol)

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
		Remote    string
		BlockTime int
		Retry     int
	}
	if err := getParamsFromJsonFile("HandshakeTimeout.json", &params); err != nil {
		_ = log4.Error("%s", err)
		return false
	}

	protocol := protocols.NewOnlyHeartbeatMsgHandler()
	setup(protocol)

	common.SetHandshakeTimeout(params.BlockTime)
	if err := ns.Connect(params.Remote); err != nil {
		log4.Debug("connecting to %s failed, err: %s", params.Remote, err)
	} else {
		log4.Info("handshake success!")
		return true
	}

	for i := 0; i < params.Retry; i++ {
		log4.Debug("connecting retry cnt %d", i)
		common.SetHandshakeTimeout(0)
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
	setup(protocol)

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
	setup(protocol)

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
	setup(protocol)

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
	setup(protocol)

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

// todo
// ddos 攻击
func DDos() bool {

	log4.Info("ddos attack end!")
	return true
}

// todo
// 异常块高
func InvalidBlockHeight() bool {
	return true
}

// todo
// 路由表攻击
func AttackRoutable() bool {
	return true
}

// todo
// 非法交易攻击
func AttackTxPool() bool {
	return true
}

// todo
// 双花
func DoubleSpend() bool {
	return true
}
