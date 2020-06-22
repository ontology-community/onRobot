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

package netserver

import (
	"errors"
	"fmt"
	"net"
	"strconv"

	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/log"
	"github.com/ontology-community/onRobot/pkg/p2pserver/common"
	"github.com/ontology-community/onRobot/pkg/p2pserver/connect_controller"
	"github.com/ontology-community/onRobot/pkg/p2pserver/mock"
	p2p "github.com/ontology-community/onRobot/pkg/p2pserver/net/protocol"
	"github.com/ontology-community/onRobot/pkg/p2pserver/peer"
	st "github.com/ontology-community/onRobot/pkg/p2pserver/stat"
)

func NewNetServerWithKid(protocol p2p.Protocol, conf *config.P2PNodeConfig, kid *common.PeerKeyId, reserveAddrFilter p2p.AddressFilter) (*NetServer, error) {
	nodePort := conf.NodePort
	if nodePort == 0 {
		nodePort = config.DEFAULT_NODE_PORT
	}

	info := peer.NewPeerInfo(kid.Id, common.PROTOCOL_VERSION, common.SERVICE_NODE, true,
		conf.HttpInfoPort, nodePort, 0, softVersion, "")

	option, err := connect_controller.ConnCtrlOptionFromConfig(conf, reserveAddrFilter)
	if err != nil {
		return nil, err
	}

	listener, err := connect_controller.NewListener(nodePort, conf)
	if err != nil {
		log.Error("[p2p]failed to create sync listener")
		return nil, errors.New("[p2p]failed to create sync listener")
	}

	log.Infof("[p2p] init peer ID to %s", info.Id.ToHexString())

	return NewCustomNetServer(kid, info, protocol, listener, option), nil
}

//NewNetServer return the net object in p2p
func NewNetServerWithTxStat(protocol p2p.Protocol, conf *config.P2PNodeConfig, reserveAddrFilter p2p.AddressFilter) (*NetServer, error) {
	nodePort := conf.NodePort
	if nodePort == 0 {
		nodePort = config.DEFAULT_NODE_PORT
	}

	keyId := common.RandPeerKeyId()
	info := peer.NewPeerInfo(keyId.Id, common.PROTOCOL_VERSION, common.SERVICE_NODE, true,
		conf.HttpInfoPort, nodePort, 0, softVersion, "")

	option, err := connect_controller.ConnCtrlOptionFromConfig(conf, reserveAddrFilter)
	if err != nil {
		return nil, err
	}

	listener, err := connect_controller.NewListener(nodePort, conf)
	if err != nil {
		log.Error("[p2p]failed to create sync listener")
		return nil, errors.New("[p2p]failed to create sync listener")
	}

	log.Infof("[p2p] init peer ID to %s", info.Id.ToHexString())

	s := NewCustomNetServer(keyId, info, protocol, listener, option)
	s.stat = st.NewMsgStat()
	s.Np = NewNbrPeersWithTxStat(s.stat)

	return s, nil
}

func NewNetServerWithSubset(listenAddr string, proto p2p.Protocol, nw mock.Network, reserves []string) *NetServer {
	const maxConn = 100

	kid := common.RandPeerKeyId()
	info := peer.NewPeerInfo(kid.Id, common.PROTOCOL_VERSION, common.SERVICE_NODE, true,
		0, 0, 0, softVersion, "")

	listener := nw.NewListenerWithAddr(kid.Id, listenAddr)
	host, port, _ := net.SplitHostPort(listenAddr)
	dialer := nw.NewDialerWithHost(kid.Id, host)
	info.Addr = listenAddr
	iport, _ := strconv.Atoi(port)
	info.Port = uint16(iport)
	opt := connect_controller.NewConnCtrlOption().MaxInBoundPerIp(maxConn).
		MaxInBound(maxConn).MaxOutBound(maxConn).WithDialer(dialer).ReservedOnly(reserves)

	return NewCustomNetServer(kid, info, proto, listener, opt)
}

func (this *NetServer) ConnectAndReturnPeer(addr string) (*peer.Peer, error) {
	peerInfo, conn, err := this.connCtrl.Connect(addr)
	if err != nil {
		return nil, err
	}
	remotePeer := peer.NewPeer(peerInfo, conn, this.NetChan)

	this.ReplacePeer(remotePeer)
	go remotePeer.Link.Rx()

	this.protocol.HandleSystemMessage(this, p2p.PeerConnected{Info: remotePeer.Info})
	return remotePeer, nil
}

func (this *NetServer) GetStat() (*st.TxStat, error) {
	if this.stat == nil {
		return nil, fmt.Errorf("stat not exist")
	}
	return this.stat, nil
}

func (this *NetServer) IsClosed() bool {
	select {
	case <-this.stopRecvCh:
		return true
	default:
	}
	return false
}

func getReservePeers(protocol p2p.Protocol, conf *config.P2PNodeConfig) p2p.AddressFilter {
	var rsv []string
	if conf.ReservedPeersOnly && conf.ReservedCfg != nil {
		rsv = conf.ReservedCfg.ReservedPeers
	}
	if rsv == nil {
		return nil
	}
	staticFilter := connect_controller.NewStaticReserveFilter(rsv)
	reserveAddrFilter := protocol.GetReservedAddrFilter(len(rsv) != 0)
	reservedPeers := p2p.CombineAddrFilter(staticFilter, reserveAddrFilter)
	return reservedPeers
}
