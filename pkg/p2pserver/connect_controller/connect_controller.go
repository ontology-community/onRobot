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
package connect_controller

import (
	"errors"
	"fmt"
	"net"
	"sort"
	"strconv"
	"sync"
	"sync/atomic"

	"github.com/ontio/ontology/common/log"
	"github.com/scylladb/go-set/strset"

	"github.com/ontology-community/onRobot/pkg/p2pserver/common"
	"github.com/ontology-community/onRobot/pkg/p2pserver/handshake"
	"github.com/ontology-community/onRobot/pkg/p2pserver/net/protocol"
	"github.com/ontology-community/onRobot/pkg/p2pserver/peer"
)

const INBOUND_INDEX = 0
const OUTBOUND_INDEX = 1

var ErrHandshakeSelf = errors.New("the node handshake with itself")

type connectedPeer struct {
	connectId uint64
	addr      string
	peer      *peer.PeerInfo
}

type ConnectController struct {
	ConnCtrlOption
	reserveAddrFilter p2p.AddressFilter //todo combine with ConnCtrlOption

	selfId   *common.PeerKeyId
	peerInfo *peer.PeerInfo

	mutex                sync.Mutex
	inoutbounds          [2]*strset.Set // in/outbounds address list
	inboundListenAddress *strset.Set    // in bound listen address
	connecting           *strset.Set
	peers                map[common.PeerId]*connectedPeer // all connected peers

	ownListenAddr string
	nextConnectId uint64
}

func NewConnectController(peerInfo *peer.PeerInfo, keyid *common.PeerKeyId,
	option ConnCtrlOption, reserveAddrFilter p2p.AddressFilter) *ConnectController {
	control := &ConnectController{
		ConnCtrlOption:       option,
		reserveAddrFilter:    reserveAddrFilter,
		selfId:               keyid,
		peerInfo:             peerInfo,
		inoutbounds:          [2]*strset.Set{strset.New(), strset.New()},
		inboundListenAddress: strset.New(),
		connecting:           strset.New(),
		peers:                make(map[common.PeerId]*connectedPeer),
	}

	// put domain to the end
	sort.Slice(control.ReservedPeers, func(i, j int) bool {
		return net.ParseIP(control.ReservedPeers[i]) != nil
	})

	return control
}

func (self *ConnectController) OwnAddress() string {
	self.mutex.Lock()
	defer self.mutex.Unlock()
	return self.ownListenAddr
}

func (self *ConnectController) SetOwnAddress(listenAddr string) {
	self.mutex.Lock()
	defer self.mutex.Unlock()
	self.ownListenAddr = listenAddr
}

func (self *ConnectController) getConnectId() uint64 {
	return atomic.AddUint64(&self.nextConnectId, 1)
}

func (self *ConnectController) hasInbound(addr string) bool {
	self.mutex.Lock()
	defer self.mutex.Unlock()
	return self.inoutbounds[INBOUND_INDEX].Has(addr)
}

func (self *ConnectController) OutboundsCount() uint {
	self.mutex.Lock()
	defer self.mutex.Unlock()
	return uint(self.inoutbounds[OUTBOUND_INDEX].Size())
}

func (self *ConnectController) InboundsCount() uint {
	self.mutex.Lock()
	defer self.mutex.Unlock()
	return uint(self.inoutbounds[INBOUND_INDEX].Size())
}

func (self *ConnectController) isBoundFull(index int) bool {
	count := self.boundsCount(index)
	if index == INBOUND_INDEX {
		return count >= self.MaxConnInBound
	}
	return count >= self.MaxConnOutBound
}

func (self *ConnectController) boundsCount(index int) uint {
	self.mutex.Lock()
	defer self.mutex.Unlock()
	return uint(self.inoutbounds[index].Size())
}

func (self *ConnectController) hasBoundAddr(addr string) bool {
	self.mutex.Lock()
	defer self.mutex.Unlock()
	has := self.inoutbounds[INBOUND_INDEX].Has(addr) || self.inoutbounds[OUTBOUND_INDEX].Has(addr)
	has = has || self.inboundListenAddress.Has(addr)
	return has
}

func (self *ConnectController) tryAddConnecting(addr string) bool {
	self.mutex.Lock()
	defer self.mutex.Unlock()

	if self.connecting.Has(addr) {
		return false
	}
	self.connecting.Add(addr)

	return true
}

func (self *ConnectController) removeConnecting(addr string) {
	self.mutex.Lock()
	defer self.mutex.Unlock()

	self.connecting.Remove(addr)
}

func (self *ConnectController) reserveEnabled() bool {
	return len(self.ReservedPeers) > 0
}

// remoteAddr format 192.168.1.1:61234
func (self *ConnectController) inReserveList(remoteIPPort string) bool {
	// 192.168.1.1 in reserve list, 192.168.1.111:61234 and 192.168.1.11:61234 can connect in if we are using prefix matching
	// so get this IP to do fully match
	remoteAddr, _, err := net.SplitHostPort(remoteIPPort)
	if err != nil {
		return false
	}
	// we don't load domain in start because we consider domain's A/AAAA record may change sometimes
	for _, curIPOrName := range self.ReservedPeers {
		curIPs, err := net.LookupHost(curIPOrName)
		if err != nil {
			continue
		}
		for _, digIP := range curIPs {
			if digIP == remoteAddr {
				return true
			}
		}
	}

	return false
}

func (self *ConnectController) checkReservedPeers(remoteAddr string) error {
	if self.reserveAddrFilter == nil {
		return nil
	}
	if !self.reserveAddrFilter.Filtered(remoteAddr) && (!self.reserveEnabled() || self.inReserveList(remoteAddr)) {
		return nil
	}
	return fmt.Errorf("local %s, the remote addr: %s not in reserved list", self.peerInfo.Addr, remoteAddr)
}

func (self *ConnectController) getInboundCountWithIp(ip string) uint {
	self.mutex.Lock()
	defer self.mutex.Unlock()
	var count uint
	self.inoutbounds[INBOUND_INDEX].Each(func(addr string) bool {
		ipRecord, _ := common.ParseIPAddr(addr)
		if ipRecord == ip {
			count += 1
		}

		return true
	})

	return count
}

func (self *ConnectController) AcceptConnect(conn net.Conn) (*peer.PeerInfo, net.Conn, error) {
	addr := conn.RemoteAddr().String()
	err := self.beforeHandshakeCheck(addr, INBOUND_INDEX)
	if err != nil {
		return nil, nil, err
	}

	peerInfo, err := handshake.HandshakeServer(self.peerInfo, self.selfId, conn)
	if err != nil {
		return nil, nil, err
	}

	err = self.afterHandshakeCheck(peerInfo, addr)
	if err != nil {
		return nil, nil, err
	}

	wrapped := self.savePeer(conn, peerInfo, INBOUND_INDEX)

	log.Infof("inbound peer %s connected, %s", conn.RemoteAddr().String(), peerInfo)
	return peerInfo, wrapped, nil
}

//Connect used to connect net address under sync or cons mode
// need call Peer.Close to clean up resource.
func (self *ConnectController) Connect(addr string) (*peer.PeerInfo, net.Conn, error) {
	err := self.beforeHandshakeCheck(addr, OUTBOUND_INDEX)
	if err != nil {
		return nil, nil, err
	}

	if !self.tryAddConnecting(addr) {
		return nil, nil, fmt.Errorf("node exist in connecting list: %s", addr)
	}
	defer self.removeConnecting(addr)

	conn, err := self.dialer.Dial(addr)
	if err != nil {
		return nil, nil, err
	}

	peerInfo, err := handshake.HandshakeClient(self.peerInfo, self.selfId, conn)
	if err != nil {
		_ = conn.Close()
		return nil, nil, err
	}

	err = self.afterHandshakeCheck(peerInfo, conn.RemoteAddr().String())
	if err != nil {
		_ = conn.Close()
		return nil, nil, err
	}

	wrapped := self.savePeer(conn, peerInfo, OUTBOUND_INDEX)

	log.Infof("outbound peer %s connected. %s", conn.RemoteAddr().String(), peerInfo)
	return peerInfo, wrapped, nil
}

func (self *ConnectController) afterHandshakeCheck(remotePeer *peer.PeerInfo, remoteAddr string) error {
	if err := self.isHandWithSelf(remotePeer, remoteAddr); err != nil {
		return err
	}

	return self.checkPeerIdAndIP(remotePeer, remoteAddr)
}

func (self *ConnectController) beforeHandshakeCheck(addr string, index int) error {
	err := self.checkReservedPeers(addr)
	if err != nil {
		return err
	}

	if self.hasBoundAddr(addr) {
		return fmt.Errorf("peer %s already in connection records", addr)
	}

	if self.OwnAddress() == addr {
		return fmt.Errorf("connecting with self address %s", addr)
	}

	if self.isBoundFull(index) {
		return fmt.Errorf("[p2p] bound %d connections reach max limit", index)
	}
	if index == INBOUND_INDEX {
		remoteIp, err := common.ParseIPAddr(addr)
		if err != nil {
			return fmt.Errorf("[p2p]parse ip error %v", err.Error())
		}
		connNum := self.getInboundCountWithIp(remoteIp)
		if connNum >= self.MaxConnInBoundPerIP {
			return fmt.Errorf("connections(%d) with ip(%s) has reach max limit(%d), "+
				"conn closed", connNum, remoteIp, self.MaxConnInBoundPerIP)
		}
	}

	return nil
}

func (self *ConnectController) isHandWithSelf(remotePeer *peer.PeerInfo, remoteAddr string) error {
	addrIp, err := common.ParseIPAddr(remoteAddr)
	if err != nil {
		log.Warn(err)
		return err
	}
	nodeAddr := addrIp + ":" + strconv.Itoa(int(remotePeer.Port))
	if remotePeer.Id.ToUint64() == self.selfId.Id.ToUint64() {
		self.SetOwnAddress(nodeAddr)
		return ErrHandshakeSelf
	}

	return nil
}

func (self *ConnectController) getPeer(kid common.PeerId) *connectedPeer {
	self.mutex.Lock()
	defer self.mutex.Unlock()
	p := self.peers[kid]
	return p
}

func (self *ConnectController) savePeer(conn net.Conn, p *peer.PeerInfo, index int) net.Conn {
	self.mutex.Lock()
	defer self.mutex.Unlock()

	addr := conn.RemoteAddr().String()
	self.inoutbounds[index].Add(addr)
	listen := p.RemoteListenAddress()
	if index == INBOUND_INDEX {
		self.inboundListenAddress.Add(listen)
	}

	cid := self.getConnectId()
	self.peers[p.Id] = &connectedPeer{
		connectId: cid,
		addr:      addr,
		peer:      p,
	}

	return &Conn{
		Conn:       conn,
		connectId:  cid,
		kid:        p.Id,
		addr:       addr,
		listenAddr: listen,
		boundIndex: index,
		controller: self,
	}
}

func (self *ConnectController) removePeer(conn *Conn) {
	self.mutex.Lock()
	defer self.mutex.Unlock()

	self.inoutbounds[conn.boundIndex].Remove(conn.addr)
	if conn.boundIndex == INBOUND_INDEX {
		self.inboundListenAddress.Remove(conn.listenAddr)
	}

	p := self.peers[conn.kid]
	if p == nil || p.peer == nil {
		log.Warnf("connection %s not in controller", conn.kid.ToHexString())
	} else if p.connectId == conn.connectId { // connection not replaced
		delete(self.peers, conn.kid)
	}
}

// if connection with peer.Kid exist, but has different IP, return error
func (self *ConnectController) checkPeerIdAndIP(peer *peer.PeerInfo, addr string) error {
	oldPeer := self.getPeer(peer.Id)
	if oldPeer == nil {
		return nil
	}

	ipOld, err := common.ParseIPAddr(oldPeer.addr)
	if err != nil {
		err := fmt.Errorf("[createPeer]exist peer ip format is wrong %s", oldPeer.addr)
		log.Fatal(err)
		return err
	}
	ipNew, err := common.ParseIPAddr(addr)
	if err != nil {
		err := fmt.Errorf("[createPeer]connecting peer ip format is wrong %s, close", addr)
		log.Fatal(err)
		return err
	}

	if ipNew != ipOld {
		err := fmt.Errorf("[createPeer]same peer id from different addr: %s, %s close latest one", ipOld, ipNew)
		log.Warn(err)
		return err
	}

	return nil
}
