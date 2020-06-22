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
	"github.com/ontio/ontology/common/log"
	msgCommon "github.com/ontology-community/onRobot/pkg/p2pserver/common"
	msgTypes "github.com/ontology-community/onRobot/pkg/p2pserver/message/types"
	p2p "github.com/ontology-community/onRobot/pkg/p2pserver/net/protocol"
)

type NeighborHandler struct {
	respch chan *msgTypes.FindNodeResp
}

func NewNeighborHandler(ch chan *msgTypes.FindNodeResp) *NeighborHandler {
	return &NeighborHandler{
		respch: ch,
	}
}

func (self *NeighborHandler) start(net p2p.P2P) {
}

func (self *NeighborHandler) stop() {
}

func (self *NeighborHandler) HandleSystemMessage(net p2p.P2P, msg p2p.SystemMessage) {
	switch m := msg.(type) {
	case p2p.NetworkStart:
		self.start(net)
	case p2p.PeerConnected:
		log.Debugf("peer connected, address: %s, id %d", m.Info.Addr, m.Info.Id.ToUint64())
	case p2p.PeerDisConnected:
		log.Debugf("peer disconnected, address: %s, id %d", m.Info.Addr, m.Info.Id.ToUint64())
	case p2p.NetworkStop:
		self.stop()
	}
}

func (self *NeighborHandler) HandlePeerMessage(ctx *p2p.Context, msg msgTypes.Message) {
	log.Trace("[p2p]receive message, remote address %s, id %d, type %s", ctx.Sender().GetAddr(), ctx.Sender().GetID().ToUint64(), msg.CmdType())
	switch m := msg.(type) {
	case *msgTypes.FindNodeResp:
		self.handleNeighborResp(ctx, m)
	case *msgTypes.NotFound:
		log.Debugf("[p2p]receive notFound message, hash is %s", m.Hash.ToHexString())
	default:
		msgType := msg.CmdType()
		if msgType == msgCommon.VERACK_TYPE || msgType == msgCommon.VERSION_TYPE {
			log.Infof("receive message: %s from peer %s", msgType, ctx.Sender().GetAddr())
		} else {
			log.Warnf("unknown message handler for the recvCh: %s", msgType)
		}
	}
}

func (self *NeighborHandler) GetReservedAddrFilter(staticFilterEnable bool) p2p.AddressFilter {
	return nil
}

func (self *NeighborHandler) handleNeighborResp(ctx *p2p.Context, trn *msgTypes.FindNodeResp) {
	self.respch <- trn
}
