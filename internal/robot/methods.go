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

package robot

import (
	log4 "github.com/alecthomas/log4go"
	ontcm "github.com/ontio/ontology/common"
	"github.com/ontology-community/onRobot/internal/robot/conf"
	"github.com/ontology-community/onRobot/pkg/files"
	p2pcm "github.com/ontology-community/onRobot/pkg/p2pserver/common"
	"github.com/ontology-community/onRobot/pkg/p2pserver/handshake"
	"github.com/ontology-community/onRobot/pkg/p2pserver/message/types"
	pr "github.com/ontology-community/onRobot/pkg/p2pserver/params"
	"github.com/ontology-community/onRobot/pkg/p2pserver/protocols"
	"github.com/ontology-community/onRobot/pkg/sdk"
	"math"
	"math/big"
	"net"
	"strconv"
	"strings"
	"time"
)

func Demo() bool {
	// get block height
	jsonrpcAddr := "http://172.168.3.158:20336"
	height, err := sdk.GetBlockCurrentHeight(jsonrpcAddr)
	if err != nil {
		_ = log4.Error("%s", err)
		return false
	}
	log4.Info("jsonrpcAddr %s current block height %d", jsonrpcAddr, height)

	// recover kp
	acc, err := sdk.RecoverAccount(conf.TransferWalletPath, conf.DefConfig.WalletPwd)
	if err != nil {
		_ = log4.Error("%s", err)
		return false
	}
	log4.Info("address %s", acc.Address.ToBase58())

	// get balance
	resp, err := sdk.GetBalance(jsonrpcAddr, acc.Address)
	if err != nil {
		_ = log4.Error("%s", err)
		return false
	}
	log4.Info("ont %s, ong %s, block height %s", resp.Ont, resp.Ong, resp.Height)
	return true
}

// 伪造peerid&随机pubkey组合成kid，由此生成netserver
// 连接后，服务端会在readmessage时校验pubkey，并通过pubkey重新生成kid，
// 所以即便连接成功了，服务端dht记录的也是根据pubkey生成的peerid
func FakePeerID() bool {
	var params struct {
		Remote       string
		DispatchTime int
	}
	if err := files.LoadParams(conf.ParamsFileDir, "FakePeerID.json", &params); err != nil {
		_ = log4.Error("%s", err)
		return false
	}

	kid := p2pcm.RandPeerKeyId()
	list, err := GenerateZeroDistancePeerIDs(kid.Id, 1)
	if err != nil {
		_ = log4.Error("%s", err)
		return false
	}
	if len(list) != 1 {
		_ = log4.Error("generate fake peer id failed")
		return false
	}
	pkid := p2pcm.FakePeerKeyId(list[0])

	protocol := protocols.NewOnlyHeartbeatMsgHandler()
	ns := GenerateNetServerWithFakeKid(protocol, pkid)
	if err := ns.Connect(params.Remote); err != nil {
		_ = log4.Error("%s", err)
	} else {
		pid := ns.GetID()
		log4.Debug("%s connected", pid.ToHexString())
	}

	Dispatch(params.DispatchTime)
	log4.Info("fake peer id success!")

	return true
}

func Connect() bool {
	// 1. get params from json file
	var params struct {
		Remote   string
		TestCase uint8
	}
	if err := files.LoadParams(conf.ParamsFileDir, "Connect.json", &params); err != nil {
		_ = log4.Error("%s", err)
		return false
	}

	// 2. set params params
	pr.SetHandshakeStopLevel(params.TestCase)

	// 3. setup p2p.protocols
	protocol := protocols.NewOnlyHeartbeatMsgHandler()
	ns := GenerateNetServerWithProtocol(protocol)

	// 4. connect and handshake
	remotePeer, err := ns.ConnectAndReturnPeer(params.Remote)
	if err != nil {
		log4.Debug("connecting to %s failed, err: %s", params.Remote, err)
		return false
	}

	// 5. calculate distance
	cpl := distance(ns.GetID(), remotePeer.GetID())

	log4.Info("handshake end success, cpl is %d", cpl)

	return true
}

func HandshakeWrongMsg() bool {

	// 1. get params from json file
	var params struct {
		Remote       string
		DispatchTime int
		Version      uint32
		Services     uint64
		TimeStamp    int64
		SyncPort     uint16
		HttpInfoPort uint16
		Nonce        uint64
		StartHeight  uint64
		Relay        uint8
		IsConsensus  bool
		SoftVersion  string
	}
	if err := files.LoadParams(conf.ParamsFileDir, "HandshakeWrongMsg.json", &params); err != nil {
		_ = log4.Error("%s", err)
		return false
	}

	pr.SetHandshakeWrongMsg(true)
	version := &types.Version{
		P: types.VersionPayload{
			Version:      params.Version,
			Services:     params.Services,
			TimeStamp:    params.TimeStamp,
			SyncPort:     params.SyncPort,
			HttpInfoPort: params.HttpInfoPort,
			Nonce:        params.Nonce,
			StartHeight:  params.StartHeight,
			Relay:        params.Relay,
			IsConsensus:  params.IsConsensus,
			SoftVersion:  params.SoftVersion,
		},
	}

	handshake.SetTestVersion(version)
	protocol := protocols.NewOnlyHeartbeatMsgHandler()
	ns := GenerateNetServerWithProtocol(protocol)

	if err := ns.Connect(params.Remote); err == nil {
		_ = log4.Error("connecting to %s with invalid version should be failed!", params.Remote)
		return false
	}

	Dispatch(params.DispatchTime)

	log4.Info("handshakeWrongMsg end!")

	return true
}

func HandshakeTimeout() bool {
	var params struct {
		Remote string
		ClientBlockTime,
		ServerBlockTime int
		Retry int
	}
	if err := files.LoadParams(conf.ParamsFileDir, "HandshakeTimeout.json", &params); err != nil {
		_ = log4.Error("%s", err)
		return false
	}

	protocol := protocols.NewOnlyHeartbeatMsgHandler()
	ns := GenerateNetServerWithProtocol(protocol)

	pr.SetHandshakeClientTimeout(params.ClientBlockTime)
	pr.SetHandshakeServerTimeout(params.ServerBlockTime)
	if err := ns.Connect(params.Remote); err != nil {
		log4.Debug("connecting to %s failed, err: %s", params.Remote, err)
	} else {
		log4.Info("handshake success!")
		return true
	}

	for i := 0; i < params.Retry; i++ {
		log4.Debug("connecting retry cnt %d", i)
		pr.SetHandshakeClientTimeout(0)
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
	if err := files.LoadParams(conf.ParamsFileDir, "Heartbeat.json", &params); err != nil {
		_ = log4.Error("%s", err)
		return false
	}

	protocol := protocols.NewOnlyHeartbeatMsgHandler()
	ns := GenerateNetServerWithProtocol(protocol)

	pr.SetHeartbeatTestBlockHeight(params.InitBlockHeight)
	if err := ns.Connect(params.Remote); err != nil {
		_ = log4.Error("connecting to %s failed, err: %s", params.Remote, err)
		return false
	}

	Dispatch(params.DispatchTime)

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
	if err := files.LoadParams(conf.ParamsFileDir, "HeartbeatInterruptPing.json", &params); err != nil {
		_ = log4.Error("%s", err)
		return false
	}

	pr.SetHeartbeatTestBlockHeight(params.InitBlockHeight)
	pr.SetHeartbeatTestInterruptAfterStartTime(params.InterruptAfterStartTime)
	pr.SetHeartbeatTestInterruptPingLastTime(params.InterruptLastTime)

	protocol := protocols.NewOnlyHeartbeatMsgHandler()
	ns := GenerateNetServerWithProtocol(protocol)

	if err := ns.Connect(params.Remote); err != nil {
		_ = log4.Error("connecting to %s failed, err: %s", params.Remote, err)
		return false
	}

	Dispatch(params.DispatchTime)

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
	if err := files.LoadParams(conf.ParamsFileDir, "HeartbeatInterruptPong.json", &params); err != nil {
		_ = log4.Error("%s", err)
		return false
	}

	pr.SetHeartbeatTestBlockHeight(params.InitBlockHeight)
	pr.SetHeartbeatTestInterruptAfterStartTime(params.InterruptAfterStartTime)
	pr.SetHeartbeatTestInterruptPongLastTime(params.InterruptLastTime)

	protocol := protocols.NewOnlyHeartbeatMsgHandler()
	ns := GenerateNetServerWithProtocol(protocol)

	if err := ns.Connect(params.Remote); err != nil {
		_ = log4.Error("connecting to %s failed, err: %s", params.Remote, err)
		return false
	}

	Dispatch(params.DispatchTime)

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
	if err := files.LoadParams(conf.ParamsFileDir, "ResetPeerID.json", &params); err != nil {
		_ = log4.Error("%s", err)
		return false
	}

	pr.SetHeartbeatTestBlockHeight(params.InitBlockHeight)
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

	Dispatch(params.DispatchTime)

	log4.Info("reset peerID end!")
	return true
}

// ddos 攻击, 构造多个peerID，连接并发送ping
// 结果: 对于同一ip，节点最多接收16个链接，其他的链接会在客户端进行连接时失败，并且不影响块同步速度
// 伪造ip的问题在于邻结点写入时从conn中获取ip，而不是从peerInfo中获取ip
func DDos() bool {

	// try to get blockheight
	var params struct {
		Remote          string
		JsonRpc         string
		InitBlockHeight uint64
		DispatchTime    int
		StartPort       int
		ConnNumber      int
	}
	if err := files.LoadParams(conf.ParamsFileDir, "DDOS.json", &params); err != nil {
		_ = log4.Error("%s", err)
		return false
	}

	height, err := sdk.GetBlockCurrentHeight(params.JsonRpc)
	if err != nil {
		_ = log4.Error("%s", err)
		return false
	} else {
		log4.Debug("block height before ddos %d", height)
	}

	pr.SetHeartbeatTestBlockHeight(params.InitBlockHeight)
	for i := 0; i < params.ConnNumber; i++ {
		port := uint16(params.StartPort + i)
		protocol := protocols.NewOnlyHeartbeatMsgHandler()
		ns := GenerateNetServerWithContinuePort(protocol, port)
		peerID := ns.GetID()
		if err := ns.Connect(params.Remote); err != nil {
			_ = log4.Error("peer %s connecting to %s failed, err: %s", peerID.ToHexString(), params.Remote, err)
		} else {
			log4.Debug("peer %s, index %d connecting to %s success", peerID.ToHexString(), int(port)-params.StartPort, params.Remote)
		}
	}

	Dispatch(params.DispatchTime)

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
	if err := files.LoadParams(conf.ParamsFileDir, "AskFakeBlocks.json", &params); err != nil {
		_ = log4.Error("%s", err)
		return false
	}

	pr.SetHeartbeatTestBlockHeight(params.InitBlockHeight)
	protocol := protocols.NewOnlyHeartbeatMsgHandler()
	ns := GenerateNetServerWithProtocol(protocol)
	peer, err := ns.ConnectAndReturnPeer(params.Remote)
	if err != nil {
		_ = log4.Error("connecting to %s failed, err: %s", params.Remote, err)
		return false
	}

	startHash, _ := sdk.Uint256FromHexString(params.StartHash)
	endHash, _ := sdk.Uint256FromHexString(params.EndHash)
	req := &types.HeadersReq{
		HashStart: startHash,
		HashEnd:   endHash,
		Len:       1,
	}
	if err = peer.Send(req); err != nil {
		_ = log4.Error("send headersReq failed, err %s", err)
		return false
	}

	// Dispatch
	if msg := protocol.Out(params.DispatchTime); msg != nil {
		log4.Debug("invalid block endHash accepted by sync node, msg %v", msg)
		return false
	} else {
		log4.Info("invalid block endHash rejected by sync node")
		return true
	}
}

// 非法交易攻击
func AttackTxPool() bool {

	// get params from json file
	var params struct {
		RemoteList               []string
		JsonRpcList              []string
		DispatchTime             int
		DestAccount              string
		TxNum                    int
		MinExpectedBlkHeightDiff uint64
	}

	if err := files.LoadParams(conf.ParamsFileDir, "AttackTxPool.json", &params); err != nil {
		_ = log4.Error("%s", err)
		return false
	}
	if len(params.RemoteList) != len(params.JsonRpcList) {
		_ = log4.Error("remote transList length != json rpc transList length")
		return false
	}

	// recover account and get balance before transfer
	acc, err := sdk.RecoverAccount(conf.TransferWalletPath, conf.DefConfig.WalletPwd)
	if err != nil {
		_ = log4.Error("%s", err)
		return false
	}

	// get balance before test
	balanceBeforeTransfer, err := GetBalanceAndCompare(params.JsonRpcList, acc)
	if err != nil || len(balanceBeforeTransfer) == 0 {
		_ = log4.Error("get balance failed")
		return false
	}

	// get block height before test
	preBlkHeightList, err := GetBlockHeightList(params.JsonRpcList)
	if err != nil || len(preBlkHeightList) != len(params.JsonRpcList) {
		_ = log4.Error("get block height list failed")
		return false
	}

	// get and set block height
	pr.SetHeartbeatTestBlockHeight(preBlkHeightList[0] + 1)

	// get peers
	peers, err := GenerateMultiHeartbeatOnlyPeers(params.RemoteList)
	if err != nil {
		_ = log4.Error("%s", err)
		return false
	}
	time.Sleep(1 * time.Second)

	// send tx
	var amount uint64 = math.MaxUint64
	transList, err := GenerateMultiOntTransfer(acc, params.DestAccount, amount, params.TxNum)
	if err != nil {
		_ = log4.Error("%s", err)
		return false
	}

	// send tx
	idx := 0
	num := params.TxNum / len(peers)
	for i := 0; i < num; i++ {
		for _, peer := range peers {
			if err := peer.Send(transList[idx]); err != nil {
				_ = log4.Warn("%s", err)
			}
			idx++
		}
	}

	// dispatch
	Dispatch(params.DispatchTime)

	// get balance after transfer
	balanceAfterTransfer, err := GetBalanceAndCompare(params.JsonRpcList, acc)
	if err != nil || len(balanceAfterTransfer) == 0 {
		_ = log4.Error("get balance failed")
		return false
	}

	// check balance
	b1, _ := new(big.Float).SetString(balanceBeforeTransfer[0].Ont)
	b2, _ := new(big.Float).SetString(balanceAfterTransfer[0].Ont)
	if b1.Cmp(b2) != 0 {
		_ = log4.Error("some invalid tx must be blocked")
		return false
	}

	// check tx
	for _, jsonrpc := range params.JsonRpcList {
		list, err := sdk.GetMemPoolTxCount(jsonrpc)
		if err != nil {
			_ = log4.Error("get %s mem pool tx count err: %s", jsonrpc, err)
		} else {
			for i, v := range list {
				log4.Debug("get %s mem pool tx count %d %d", jsonrpc, i, v)
			}
		}
	}
	//for _, jsonrpc := range params.JsonRpcList {
	//	for _, tx := range transList {
	//		hash := tx.Txn.Hash()
	//		tx, err := params.GetTxByHash(jsonrpc, hash)
	//		if tx != nil || err == nil {
	//			_ = log4.Warn("invalid tx persisted in txn pool")
	//		} else {
	//			log4.Debug("node %s txnpool without tx %s", jsonrpc, hash.ToHexString())
	//		}
	//	}
	//}

	// get current block height
	curBlkHeightList, err := GetBlockHeightList(params.JsonRpcList)
	if err != nil || len(curBlkHeightList) != len(params.JsonRpcList) {
		_ = log4.Error("get block height list failed")
		return false
	}

	// check block height
	for i, node := range params.JsonRpcList {
		pre := preBlkHeightList[i]
		cur := curBlkHeightList[i]
		dif := cur - pre
		if dif < params.MinExpectedBlkHeightDiff {
			_ = log4.Error("node %s, block height %d < %d", node, dif, params.MinExpectedBlkHeightDiff)
			return false
		} else {
			log4.Info("current block height %d, pre block height %d, diff %d", cur, pre, dif)
		}
	}

	return true
}

// 双花
func DoubleSpend() bool {
	var params struct {
		RemoteList   []string
		JsonRpcList  []string
		DispatchTime int
		DestAccount  string
	}
	if err := files.LoadParams(conf.ParamsFileDir, "DoubleSpend.json", &params); err != nil {
		_ = log4.Error("%s", err)
		return false
	}
	if len(params.RemoteList) != len(params.JsonRpcList) {
		_ = log4.Error("remote list length != json rpc list length")
		return false
	}

	// recover account and get balance before transfer
	acc, err := sdk.RecoverAccount(conf.TransferWalletPath, conf.DefConfig.WalletPwd)
	if err != nil {
		_ = log4.Error("%s", err)
		return false
	}

	balanceBeforeTransfer, err := GetBalanceAndCompare(params.JsonRpcList, acc)
	if err != nil || len(balanceBeforeTransfer) == 0 {
		_ = log4.Error("get balance failed")
		return false
	}

	// get and set block height
	if _, err := GetAndSetBlockHeight(params.JsonRpcList[0], 1); err != nil {
		_ = log4.Error("%s", err)
		return false
	}

	// get peer
	peers, err := GenerateMultiHeartbeatOnlyPeers(params.RemoteList)
	if err != nil {
		_ = log4.Error("%s", err)
		return false
	}

	amount, err := strconv.Atoi(balanceBeforeTransfer[0].Ont)
	if err != nil {
		_ = log4.Error("%s", err)
		return false
	}
	log4.Info("=====%s balance %s, and will transfer %d ont", acc.Address.ToBase58(), balanceBeforeTransfer[0].Ont, amount)

	// send preTx
	transList := make([]*types.Trn, 0, len(peers))
	for _, peer := range peers {
		tran, err := GenerateTransferOntTx(acc, params.DestAccount, uint64(amount))
		if err != nil {
			_ = log4.Error("%s", err)
			return false
		}
		if err := peer.Send(tran); err != nil {
			_ = log4.Error("%s", err)
			return false
		}
		transList = append(transList, tran)
	}

	// dispatch
	Dispatch(params.DispatchTime)

	// get and check balance after transfer
	balanceAfterTransfer, err := GetBalanceAndCompare(params.JsonRpcList, acc)
	if err != nil || len(balanceAfterTransfer) == 0 {
		_ = log4.Error("get balance failed")
		return false
	}
	if balanceAfterTransfer[0].Ont != "0" {
		_ = log4.Error("balance after transfer should be 0")
		return false
	}

	// check preTx
	succeed := make(map[string]struct{})
	for _, pretx := range transList {
		prehash := pretx.Txn.Hash()
		for _, jsonrpc := range params.JsonRpcList {
			curtx, err := sdk.GetTxByHash(jsonrpc, prehash)
			if err != nil {
				_ = log4.Error("===== node %s, failed preTx %s", jsonrpc, prehash.ToHexString())
			} else {
				safehash := strings.ToLower(prehash.ToHexString())
				succeed[safehash] = struct{}{}

				var bz []byte
				sink := ontcm.NewZeroCopySink(bz)
				payload := curtx.Payload
				payload.Serialization(sink)
				log4.Debug("===== node %s, succeed preTx %s, payload %s", jsonrpc, prehash.ToHexString(), string(bz))
			}
		}
	}
	if len(succeed) > 0 {
		_ = log4.Warn("===== more than 1 tx in txpool")
	}

	return true
}

func TransferOnt() bool {
	var params struct {
		Remote       string
		JsonRpc      string
		DispatchTime int
		DestAccount  string
		Amount       uint64
	}

	if err := files.LoadParams(conf.ParamsFileDir, "Transfer.json", &params); err != nil {
		_ = log4.Error("%s", err)
		return false
	}
	err := singleTransfer(params.Remote, params.JsonRpc, params.DestAccount, params.Amount, params.DispatchTime)
	if err != nil {
		_ = log4.Error("%s", err)
		return false
	}

	return true
}

func TxCount() bool {
	var params struct {
		IpList        []string
		StartHttpPort uint16
		EndHttpPort   uint16
		Remote        string
		DestAccount   string
		DispatchTime  int
		Ticker        int
	}

	if err := files.LoadParams(conf.ParamsFileDir, "TxCount.json", &params); err != nil {
		_ = log4.Error("%s", err)
		return false
	}

	// clear history data
	cli := NewHttpClient()
	cli.clearMsgCount(params.IpList, params.StartHttpPort, params.EndHttpPort)

	// send invalid transfer
	worker, err := NewInvalidTxWorker(params.Remote)
	if err != nil {
		_ = log4.Error("%s", err)
		return false
	}
	ticker := time.NewTicker(time.Duration(params.Ticker) * time.Second)
	dispatch := 0
	for dispatch < params.DispatchTime {
		select {
		case <-ticker.C:
			if err := worker.sendInvalidTxWithoutCheckBalance(params.DestAccount); err != nil {
				_ = log4.Error("%s", err)
				continue
			}
			dispatch += 1
		}
	}

	// get stat and dump
	list := cli.statMsgCount(params.IpList, params.StartHttpPort, params.EndHttpPort)
	totalSend, totalRecv := uint64(0), uint64(0)
	for _, v := range list {
		totalSend += v.Send
		totalRecv += v.Recv
		log4.Info("send amount %d, recv amount %d", v.Send, v.Recv)
	}
	avSend := float64(totalSend) / float64(len(list))
	avRecv := float64(totalRecv) / float64(len(list))
	log4.Info("avg send %f, avg recv %f, total send %d, total recv %d", avSend, avRecv, totalSend, totalRecv)
	return true
}

// 访问邻结点
func Neighbor() bool {
	var params struct {
		Remote         string
		ExpectedIpList []string
		Timeout        int
	}

	if err := files.LoadParams(conf.ParamsFileDir, "Neighbor.json", &params); err != nil {
		_ = log4.Error("%s", err)
		return false
	}

	ch := make(chan *types.FindNodeResp)
	protocol := protocols.NewNeighborHandler(ch)
	ns := GenerateNetServerWithProtocol(protocol)
	pr, err := ns.ConnectAndReturnPeer(params.Remote)
	if err != nil {
		_ = log4.Error("connect peer err: %s", err)
		return false
	}

	ticker := time.NewTicker(1 * time.Second)
	count := 0

	// 寻找离自己较近的节点
	req := &types.FindNodeReq{TargetID: ns.GetID()}
	nbrs := make(map[string]struct{})
	for _, v := range params.ExpectedIpList {
		nbrs[v] = struct{}{}
	}

	for {
		select {
		case msg := <-ch:
			for _, peer := range msg.CloserPeers {
				ip, _, err := net.SplitHostPort(peer.Address)
				if err != nil {
					_ = log4.Error("splitHostPort err: %s", err)
					return false
				}
				log4.Info("neighbor: %s", ip)

				if _, ok := nbrs[ip]; ok {
					delete(nbrs, ip)
				}
				if len(nbrs) == 0 {
					return true
				}
			}

		case <-ticker.C:
			if err := pr.Send(req); err != nil {
				_ = log4.Error("send find neighbor nodes req err %s", err)
				return false
			}
			if count++; count > params.Timeout {
				_ = log4.Error("testcase timeout")
				return false
			}
		}
	}
}
