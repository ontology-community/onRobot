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

package protocols

import (
	"strconv"

	log4 "github.com/alecthomas/log4go"
	"github.com/ontio/ontology/common/config"
	msgCommon "github.com/ontology-community/onRobot/pkg/p2pserver/common"
	msgTypes "github.com/ontology-community/onRobot/pkg/p2pserver/message/types"
	"github.com/ontology-community/onRobot/pkg/p2pserver/net/protocol"
	"github.com/ontology-community/onRobot/pkg/p2pserver/protocols/bootstrap"
	"github.com/ontology-community/onRobot/pkg/p2pserver/protocols/discovery"
	"github.com/ontology-community/onRobot/pkg/p2pserver/protocols/heatbeat"
	"github.com/ontology-community/onRobot/pkg/p2pserver/protocols/recent_peers"
	"github.com/ontology-community/onRobot/pkg/p2pserver/protocols/reconnect"
)

type WithoutBlockSyncMsgHandler struct {
	reconnect                *reconnect.ReconnectService
	discovery                *discovery.Discovery
	heatBeat                 *heatbeat.HeartBeat
	bootstrap                *bootstrap.BootstrapService
	persistRecentPeerService *recent_peers.PersistRecentPeerService
}

func NewWithoutBlockSyncMsgHandler() *WithoutBlockSyncMsgHandler {
	return &WithoutBlockSyncMsgHandler{}
}

func (self *WithoutBlockSyncMsgHandler) start(net p2p.P2P) {
	self.reconnect = reconnect.NewReconectService(net)
	self.discovery = discovery.NewDiscovery(net, config.DefConfig.P2PNode.ReservedCfg.MaskPeers, 0)
	seeds := config.DefConfig.Genesis.SeedList
	self.bootstrap = bootstrap.NewBootstrapService(net, seeds)
	// mark:
	self.heatBeat = heatbeat.NewHeartBeat(net)
	self.persistRecentPeerService = recent_peers.NewPersistRecentPeerService(net)
	go self.persistRecentPeerService.Start()
	go self.reconnect.Start()
	go self.discovery.Start()
	go self.heatBeat.Start()
	go self.bootstrap.Start()
}

func (self *WithoutBlockSyncMsgHandler) stop() {
	self.reconnect.Stop()
	self.discovery.Stop()
	self.persistRecentPeerService.Stop()
	self.heatBeat.Stop()
	self.bootstrap.Stop()
}

func (self *WithoutBlockSyncMsgHandler) HandleSystemMessage(net p2p.P2P, msg p2p.SystemMessage) {
	switch m := msg.(type) {
	case p2p.NetworkStart:
		self.start(net)
	case p2p.PeerConnected:
		self.reconnect.OnAddPeer(m.Info)
		self.discovery.OnAddPeer(m.Info)
		self.bootstrap.OnAddPeer(m.Info)
		self.persistRecentPeerService.AddNodeAddr(m.Info.Addr + strconv.Itoa(int(m.Info.Port)))
	case p2p.PeerDisConnected:
		self.reconnect.OnDelPeer(m.Info)
		self.discovery.OnDelPeer(m.Info)
		self.bootstrap.OnDelPeer(m.Info)
		self.persistRecentPeerService.DelNodeAddr(m.Info.Addr + strconv.Itoa(int(m.Info.Port)))
	case p2p.NetworkStop:
		self.stop()
	}
}

func (self *WithoutBlockSyncMsgHandler) HandlePeerMessage(ctx *p2p.Context, msg msgTypes.Message) {
	pid := ctx.Sender().GetID()

	switch m := msg.(type) {
	case *msgTypes.AddrReq:
		self.discovery.AddrReqHandle(ctx)
	case *msgTypes.FindNodeResp:
		self.discovery.FindNodeResponseHandle(ctx, m)
	case *msgTypes.FindNodeReq:
		self.discovery.FindNodeHandle(ctx, m)
	case *msgTypes.HeadersReq:
		HeadersReqHandle(ctx, m)
	case *msgTypes.Ping:
		self.heatBeat.PingHandle(ctx, m)
	case *msgTypes.Pong:
		self.heatBeat.PongHandle(ctx, m)
	case *msgTypes.Consensus:
		ConsensusHandle(ctx, m)
	case *msgTypes.Trn:
		log4.Trace("[p2p-without block sync] receive message, type:%s, sender: %s, send pid: %s",
			msg.CmdType(), ctx.Sender().GetAddr(), pid.ToHexString())

		TransactionHandle(ctx, m)
	case *msgTypes.Addr:
		self.discovery.AddrHandle(ctx, m)
	case *msgTypes.DataReq:
		DataReqHandle(ctx, m)
	case *msgTypes.Inv:
		InvHandle(ctx, m)
	case *msgTypes.NotFound:
		log4.Debug("[p2p]receive notFound message, hash is ", m.Hash)
	default:
		msgType := msg.CmdType()
		if msgType == msgCommon.VERACK_TYPE || msgType == msgCommon.VERSION_TYPE {
			log4.Info("receive message: %s from peer %s", msgType, ctx.Sender().GetAddr())
		} else {
			log4.Warn("unknown message handler for the recvCh: ", msgType)
		}
	}
}

func (mh *WithoutBlockSyncMsgHandler) ReconnectService() *reconnect.ReconnectService {
	return mh.reconnect
}
