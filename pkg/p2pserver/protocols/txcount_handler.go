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
	"fmt"
	"github.com/ontio/ontology/account"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/log"
	"github.com/ontology-community/onRobot/pkg/p2pserver/protocols/utils"
	"strconv"

	p2p "github.com/ontology-community/onRobot/pkg/p2pserver/net/protocol"
	"github.com/ontology-community/onRobot/pkg/p2pserver/protocols/subnet"

	"github.com/ontology-community/onRobot/pkg/p2pserver/common"
	msgTypes "github.com/ontology-community/onRobot/pkg/p2pserver/message/types"
	"github.com/ontology-community/onRobot/pkg/p2pserver/protocols/bootstrap"
	"github.com/ontology-community/onRobot/pkg/p2pserver/protocols/discovery"
	"github.com/ontology-community/onRobot/pkg/p2pserver/protocols/heatbeat"
	"github.com/ontology-community/onRobot/pkg/p2pserver/protocols/recent_peers"
	"github.com/ontology-community/onRobot/pkg/p2pserver/protocols/reconnect"
)

type TxCountHandler struct {
	seeds *utils.HostsResolver

	reconnect                *reconnect.ReconnectService
	discovery                *discovery.Discovery
	heatBeat                 *heatbeat.HeartBeat
	bootstrap                *bootstrap.BootstrapService
	persistRecentPeerService *recent_peers.PersistRecentPeerService
	subnet                   *subnet.SubNet
}

func NewTxCountHandler(acc *account.Account) *TxCountHandler {
	m := &TxCountHandler{}

	seedsList := config.DefConfig.Genesis.SeedList
	seeds, invalid := utils.NewHostsResolver(seedsList)
	if invalid != nil {
		panic(fmt.Errorf("invalid seed list； %v", invalid))
	}
	gov := utils.NewGovNodeMockResolver(nil)

	m.seeds = seeds
	m.subnet = subnet.NewSubNet(acc, seeds, gov)

	return m
}

func (self *TxCountHandler) start(net p2p.P2P) {

	self.reconnect = reconnect.NewReconectService(net)
	maskFilter := self.subnet.GetMaskAddrFilter()
	self.discovery = discovery.NewDiscovery(net, config.DefConfig.P2PNode.ReservedCfg.MaskPeers, maskFilter, 0)
	self.bootstrap = bootstrap.NewBootstrapService(net, self.seeds)

	// mark:
	self.heatBeat = heatbeat.NewHeartBeat(net)
	self.persistRecentPeerService = recent_peers.NewPersistRecentPeerService(net)
	go self.persistRecentPeerService.Start()
	go self.reconnect.Start()
	go self.discovery.Start()
	go self.heatBeat.Start()
	go self.bootstrap.Start()
	go self.subnet.Start(net)
}

func (self *TxCountHandler) stop() {
	self.reconnect.Stop()
	self.discovery.Stop()
	self.persistRecentPeerService.Stop()
	self.heatBeat.Stop()
	self.bootstrap.Stop()
	self.subnet.Stop()
}

func (self *TxCountHandler) HandleSystemMessage(net p2p.P2P, msg p2p.SystemMessage) {
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
	case p2p.HostAddrDetected:
		self.subnet.OnHostAddrDetected(m.ListenAddr)
	}
}

func (self *TxCountHandler) HandlePeerMessage(ctx *p2p.Context, msg msgTypes.Message) {
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
		log.Tracef("[p2p-without block sync] receive message, type:%s, sender: %s, send pid: %s",
			msg.CmdType(), ctx.Sender().GetAddr(), pid.ToHexString())
		TransactionHandle(ctx, m)
	case *msgTypes.Addr:
		self.discovery.AddrHandle(ctx, m)
	case *msgTypes.DataReq:
		DataReqHandle(ctx, m)
	case *msgTypes.Inv:
		InvHandle(ctx, m)
	case *msgTypes.SubnetMembersRequest:
		self.subnet.OnMembersRequest(ctx, m)
	case *msgTypes.SubnetMembers:
		self.subnet.OnMembersResponse(ctx, m)
	case *msgTypes.NotFound:
		log.Debugf("[p2p]receive notFound message, hash is ", m.Hash)
	default:
		msgType := msg.CmdType()
		if msgType == common.VERACK_TYPE || msgType == common.VERSION_TYPE {
			log.Infof("receive message: %s from peer %s", msgType, ctx.Sender().GetAddr())
		} else {
			log.Warnf("unknown message handler for the recvCh: ", msgType)
		}
	}
}

func (self *TxCountHandler) GetReservedAddrFilter() p2p.AddressFilter {
	return self.subnet.GetReservedAddrFilter()
}
