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
	"time"

	"github.com/ontio/ontology/common/log"

	msgCommon "github.com/ontology-community/onRobot/pkg/p2pserver/common"
	msgTypes "github.com/ontology-community/onRobot/pkg/p2pserver/message/types"
	p2p "github.com/ontology-community/onRobot/pkg/p2pserver/net/protocol"
	"github.com/ontology-community/onRobot/pkg/p2pserver/protocols/heatbeat"
)

type RemoteMessage struct {
	Context *p2p.Context
	Message msgTypes.Message
}

type OnlyHeartbeatMsgHandler struct {
	heatBeat *heatbeat.HeartBeat
	recvCh   chan *RemoteMessage
}

func NewOnlyHeartbeatMsgHandler() *OnlyHeartbeatMsgHandler {
	return &OnlyHeartbeatMsgHandler{
		recvCh: make(chan *RemoteMessage, 100),
	}
}

func (self *OnlyHeartbeatMsgHandler) start(net p2p.P2P) {
	self.heatBeat = heatbeat.NewHeartBeat(net)
	go self.heatBeat.Start()
}

func (self *OnlyHeartbeatMsgHandler) stop() {
	self.heatBeat.Stop()
}

func (self *OnlyHeartbeatMsgHandler) HandleSystemMessage(net p2p.P2P, msg p2p.SystemMessage) {
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

func (self *OnlyHeartbeatMsgHandler) HandlePeerMessage(ctx *p2p.Context, msg msgTypes.Message) {
	log.Tracef("[p2p]receive message, remote address %s, id %d, type %s", ctx.Sender().GetAddr(), ctx.Sender().GetID().ToUint64(), msg.CmdType())
	switch m := msg.(type) {
	case *msgTypes.Ping:
		self.heatBeat.PingHandle(ctx, m)
	case *msgTypes.Pong:
		self.heatBeat.PongHandle(ctx, m)
	case *msgTypes.BlkHeader:
		self.BlockHeaderHandler(ctx, m)
	case *msgTypes.Block:
		self.BlockHandler(ctx, m)
	case *msgTypes.Trn:
		self.TranHandler(ctx, m)
	case *msgTypes.NotFound:
		log.Debugf("[p2p]receive notFound message, hash is %s", m.Hash.ToHexString())
	default:
		msgType := msg.CmdType()
		if msgType == msgCommon.VERACK_TYPE || msgType == msgCommon.VERSION_TYPE {
			log.Infof("receive message: %s from peer %s", msgType, ctx.Sender().GetAddr())
		}
	}
}

func (self *OnlyHeartbeatMsgHandler) GetReservedAddrFilter() p2p.AddressFilter {
	return nil
}

func (self *OnlyHeartbeatMsgHandler) BlockHeaderHandler(ctx *p2p.Context, msg *msgTypes.BlkHeader) {
	self.recvCh <- &RemoteMessage{
		Context: ctx,
		Message: msg,
	}
}

func (self *OnlyHeartbeatMsgHandler) BlockHandler(ctx *p2p.Context, msg *msgTypes.Block) {
	self.recvCh <- &RemoteMessage{
		Context: ctx,
		Message: msg,
	}
}

func (self *OnlyHeartbeatMsgHandler) TranHandler(ctx *p2p.Context, msg *msgTypes.Trn) {
}

func (self *OnlyHeartbeatMsgHandler) Out(sec int) *RemoteMessage {
	timer := time.NewTimer(time.Second * time.Duration(sec))
	select {
	case msg := <-self.recvCh:
		return msg
	case <-timer.C:
		break
	}
	return nil
}
