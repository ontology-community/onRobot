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
package heatbeat

import (
	"sync/atomic"
	"time"

	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/log"

	"github.com/ontology-community/onRobot/pkg/p2pserver/common"
	msgpack "github.com/ontology-community/onRobot/pkg/p2pserver/message/msg_pack"
	"github.com/ontology-community/onRobot/pkg/p2pserver/message/types"
	p2p "github.com/ontology-community/onRobot/pkg/p2pserver/net/protocol"
	"github.com/ontology-community/onRobot/pkg/p2pserver/params"
)

type HeartBeat struct {
	net    p2p.P2P
	id     common.PeerId
	quit   chan bool
	height uint64
	start  int64
}

func NewHeartBeat(net p2p.P2P) *HeartBeat {
	return &HeartBeat{
		id:     net.GetID(),
		net:    net,
		quit:   make(chan bool),
		height: params.HeartbeatBlockHeight,
		start:  time.Now().Unix(),
	}
}

func (self *HeartBeat) Start() {
	go self.heartBeatService()
}

func (self *HeartBeat) Stop() {
	if !self.IsClosed() {
		close(self.quit)
	}
}

func (self *HeartBeat) IsClosed() bool {
	select {
	case <-self.quit:
		return true
	default:
	}
	return false
}

func (this *HeartBeat) heartBeatService() {
	var periodTime uint = config.DEFAULT_GEN_BLOCK_TIME / common.UPDATE_RATE_PER_BLOCK
	t := time.NewTicker(time.Second * (time.Duration(periodTime)))

	for {
		select {
		case <-t.C:
			this.ping()
			this.timeout()
		case <-this.quit:
			t.Stop()
			return
		}
	}
}

func (this *HeartBeat) ping() {
	if this.NeedInterrupt(true) {
		log.Debug("[p2p]interrupt ping...")
		return
	}

	height := this.height
	ping := msgpack.NewPingMsg(uint64(height))
	go this.net.Broadcast(ping)

	// log.Debugf("[p2p]send ping msg height %d", height)
}

//timeout trace whether some peer be long time no response
func (this *HeartBeat) timeout() {
	peers := this.net.GetNeighbors()
	var periodTime uint = config.DEFAULT_GEN_BLOCK_TIME / common.UPDATE_RATE_PER_BLOCK
	for _, p := range peers {
		t := p.GetContactTime()
		if t.Before(time.Now().Add(-1 * time.Second *
			time.Duration(periodTime) * common.KEEPALIVE_TIMEOUT)) {
			log.Warnf("[p2p]keep alive timeout!!!lost remote peer %d - %s from %s", p.GetID(), p.Link.GetAddr(), t.String())
			p.Close()
		}
	}
}

func (this *HeartBeat) PingHandle(ctx *p2p.Context, ping *types.Ping) {
	// mark:
	if this.NeedInterrupt(false) {
		log.Info("[p2p]interrupt pong...")
		return
	}

	remotePeer := ctx.Sender()
	remotePeer.SetHeight(ping.Height)
	p2p := ctx.Network()

	height := this.height
	p2p.SetHeight(uint64(height))
	msg := msgpack.NewPongMsg(uint64(height))

	err := remotePeer.Send(msg)
	if err != nil {
		log.Warn(err)
	}
}

func (this *HeartBeat) PongHandle(ctx *p2p.Context, pong *types.Pong) {
	remotePeer := ctx.Network()
	remotePeer.SetHeight(pong.Height)
	// mark:
	atomic.AddUint64(&this.height, 1)
}

func (this *HeartBeat) NeedInterrupt(iscli bool) bool {
	lastTime := params.HeartbeatInterruptPingLastTime
	if !iscli {
		lastTime = params.HeartbeatInterruptPongLastTime
	}

	breakAfterStart := params.HeartbeatInterruptAfterStartTime
	now := time.Now().Unix()
	start := this.start + breakAfterStart
	end := this.start + breakAfterStart + lastTime

	if breakAfterStart > 0 && now > start && now < end {
		return true
	}
	return false
}
