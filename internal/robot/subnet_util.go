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
	"net"
	"strconv"

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/account"
	"github.com/ontology-community/onRobot/internal/robot/conf"

	p2pcm "github.com/ontology-community/onRobot/pkg/p2pserver/common"
	"github.com/ontology-community/onRobot/pkg/p2pserver/mock"
	"github.com/ontology-community/onRobot/pkg/p2pserver/net/netserver"
	"github.com/ontology-community/onRobot/pkg/p2pserver/net/protocol"
	"github.com/ontology-community/onRobot/pkg/p2pserver/protocols"
	"github.com/ontology-community/onRobot/pkg/p2pserver/protocols/utils"
)

func generateNetServerWithSubnet(govPubKeys []keypair.PublicKey, acc *account.Account,
	seeds []string, host string, nw mock.Network) *netserver.NetServer {

	resolver := utils.NewGovNodeMockResolver()
	addMultiGovNodes(resolver, govPubKeys)
	protocol := protocols.NewSubnetHandler(acc, seeds, resolver)
	node := netserver.NewNetServerWithSubset(host, protocol, nw)
	return node
}

func generateMultiPubkeys(n int) ([]keypair.PublicKey, []*account.Account) {
	accList := make([]*account.Account, n)
	pubkeyList := make([]keypair.PublicKey, n)
	for i := 0; i < n; i++ {
		acc := account.NewAccount("")
		accList[i] = acc
		pubkeyList[i] = acc.PublicKey
	}
	return pubkeyList, accList
}

type nodeType uint

const (
	nodeTypeSeed nodeType = iota
	nodeTypeGov
	nodeTypeNorm
)

type wrapNode struct {
	node     *netserver.NetServer
	acc      *account.Account
	host     string
	nodeType nodeType
	seeds    []string
	gov      []keypair.PublicKey
}

func GenerateSeedNode(gov []keypair.PublicKey, seeds []string, host string, nw mock.Network) *wrapNode {
	wn := &wrapNode{
		gov:      gov,
		seeds:    seeds,
		acc:      account.NewAccount(""),
		host:     host,
		nodeType: nodeTypeSeed,
	}
	wn.node = generateNetServerWithSubnet(wn.gov, wn.acc, wn.seeds, wn.host, nw)
	return wn
}

func GenerateNormNode(gov []keypair.PublicKey, seeds []string, host string, nw mock.Network) *wrapNode {
	wn := &wrapNode{
		gov:      gov,
		seeds:    seeds,
		acc:      account.NewAccount(""),
		host:     host,
		nodeType: nodeTypeNorm,
	}
	wn.node = generateNetServerWithSubnet(wn.gov, wn.acc, wn.seeds, wn.host, nw)
	return wn
}

func GenerateGovNode(gov []keypair.PublicKey, seeds []string, host string, acc *account.Account, nw mock.Network) *wrapNode {
	wn := &wrapNode{
		gov:      gov,
		seeds:    seeds,
		acc:      acc,
		host:     host,
		nodeType: nodeTypeGov,
	}
	wn.node = generateNetServerWithSubnet(wn.gov, wn.acc, wn.seeds, wn.host, nw)
	return wn
}

func addMultiGovNodes(resolver utils.GovNodeResolver, kps []keypair.PublicKey) {
	for _, kp := range kps {
		resolver.AddGovNode(kp)
	}
}

func getSubnetMemberInfo(protocol p2p.Protocol) []p2pcm.SubnetMemberInfo {
	handler, ok := protocol.(*protocols.SubnetHandler)
	if !ok {
		return nil
	}

	return handler.GetSubnetMembersInfo()
}

func subnetDefConfig(addr string) error {
	_, port, err := net.SplitHostPort(addr)
	if err != nil {
		return err
	}
	iport, err := strconv.Atoi(port)
	if err != nil {
		return err
	}
	conf.DefConfig.Net.ReservedPeersOnly = false
	//conf.DefConfig.Net.ReservedCfg.ReservedPeers = nil
	conf.DefConfig.Net.NodePort = uint16(iport)
	conf.DefConfig.Net.HttpInfoPort = 0
	conf.DefConfig.Net.MaxConnInBound = 100
	conf.DefConfig.Net.MaxConnOutBound = 100
	return nil
}
