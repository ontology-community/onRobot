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
	"time"

	"github.com/ontio/ontology/account"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/log"

	"github.com/ontology-community/onRobot/pkg/p2pserver/common"
	msgTypes "github.com/ontology-community/onRobot/pkg/p2pserver/message/types"
	p2p "github.com/ontology-community/onRobot/pkg/p2pserver/net/protocol"
	"github.com/ontology-community/onRobot/pkg/p2pserver/protocols/bootstrap"
	"github.com/ontology-community/onRobot/pkg/p2pserver/protocols/discovery"
	"github.com/ontology-community/onRobot/pkg/p2pserver/protocols/reconnect"
	"github.com/ontology-community/onRobot/pkg/p2pserver/protocols/subnet"
	"github.com/ontology-community/onRobot/pkg/p2pserver/protocols/utils"
)

type SubnetHandler struct {
	seeds     *utils.HostsResolver
	discovery *discovery.Discovery
	bootstrap *bootstrap.BootstrapService
	subnet    *subnet.SubNet
	acct      *account.Account // nil if conenesus is not enabled
}

func NewSubnetHandler(acct *account.Account, seedList []string, ledger *utils.MockLedger) *SubnetHandler {
	gov := utils.NewGovNodeMockResolver(ledger)
	seeds, invalid := utils.NewHostsResolver(seedList)
	if invalid != nil {
		panic(fmt.Errorf("invalid seed list； %v", invalid))
	}
	subNet := subnet.NewSubNet(acct, seeds, gov)
	return &SubnetHandler{seeds: seeds, subnet: subNet, acct: acct}
}

func (self *SubnetHandler) start(net p2p.P2P) {
	maskFilter := self.subnet.GetMaskAddrFilter()
	self.discovery = discovery.NewDiscovery(net, config.DefConfig.P2PNode.ReservedCfg.MaskPeers,
		maskFilter, time.Millisecond*1000)
	self.bootstrap = bootstrap.NewBootstrapService(net, self.seeds)
	go self.discovery.Start()
	go self.bootstrap.Start()
	go self.subnet.Start(net)
}

func (self *SubnetHandler) stop() {
	self.discovery.Stop()
	self.bootstrap.Stop()
	self.subnet.Stop()
}

func (self *SubnetHandler) HandleSystemMessage(net p2p.P2P, msg p2p.SystemMessage) {
	switch m := msg.(type) {
	case p2p.NetworkStart:
		self.start(net)
	case p2p.PeerConnected:
		self.discovery.OnAddPeer(m.Info)
		self.bootstrap.OnAddPeer(m.Info)
		self.subnet.OnAddPeer(net, m.Info)
	case p2p.PeerDisConnected:
		self.discovery.OnDelPeer(m.Info)
		self.bootstrap.OnDelPeer(m.Info)
		self.subnet.OnDelPeer(m.Info)
	case p2p.NetworkStop:
		self.stop()
	case p2p.HostAddrDetected:
		self.subnet.OnHostAddrDetected(m.ListenAddr)
	}
}

func (self *SubnetHandler) HandlePeerMessage(ctx *p2p.Context, msg msgTypes.Message) {
	log.Trace("[p2p]receive message", ctx.Sender().GetAddr(), ctx.Sender().GetID())
	switch m := msg.(type) {
	case *msgTypes.AddrReq:
		self.discovery.AddrReqHandle(ctx)
	case *msgTypes.FindNodeResp:
		self.discovery.FindNodeResponseHandle(ctx, m)
	case *msgTypes.FindNodeReq:
		self.discovery.FindNodeHandle(ctx, m)
	case *msgTypes.Addr:
		self.discovery.AddrHandle(ctx, m)
	case *msgTypes.SubnetMembersRequest:
		self.subnet.OnMembersRequest(ctx, m)
	case *msgTypes.SubnetMembers:
		self.subnet.OnMembersResponse(ctx, m)
	case *msgTypes.NotFound:
		log.Debugf("[p2p]receive notFound message, hash is ", m.Hash)
	default:
		msgType := msg.CmdType()
		log.Warnf("unknown message handler for the msg: ", msgType)
	}
}

func (self *SubnetHandler) GetReservedAddrFilter(staticFilterEnabled bool) p2p.AddressFilter {
	return self.subnet.GetReservedAddrFilter(staticFilterEnabled)
}

func (self *SubnetHandler) GetMaskAddrFilter() p2p.AddressFilter {
	return self.subnet.GetMaskAddrFilter()
}

func (self *SubnetHandler) ReconnectService() *reconnect.ReconnectService {
	return nil
}

func (self *SubnetHandler) GetSubnetMembersInfo() []common.SubnetMemberInfo {
	return self.subnet.GetMembersInfo()
}
