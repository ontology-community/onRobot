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
	"fmt"
	"net"
	"strconv"

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/account"
	"github.com/ontology-community/onRobot/internal/robot/conf"
	p2pcm "github.com/ontology-community/onRobot/pkg/p2pserver/common"
	"github.com/ontology-community/onRobot/pkg/p2pserver/net/netserver"
	p2p "github.com/ontology-community/onRobot/pkg/p2pserver/net/protocol"
	"github.com/ontology-community/onRobot/pkg/p2pserver/protocols"
	"github.com/ontology-community/onRobot/pkg/p2pserver/protocols/utils"
)

func GenerateNetServerWithSubnet(govPubKeys []keypair.PublicKey, acc *account.Account, seeds []string, host string) (*netserver.NetServer, error) {
	if err := settleDefConfigPort(host); err != nil {
		return nil, err
	}
	resolver := utils.NewGovNodeMockResolver()
	addMultiGovNodes(resolver, govPubKeys)
	protocol := protocols.NewSubnetHandler(acc, seeds, resolver)
	node, err := netserver.NewNetServer(protocol, conf.DefConfig.Net)
	return node, err
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

func generateSeedNode(gov []keypair.PublicKey, seeds []string, host string) *wrapNode {
	return &wrapNode{
		gov:      gov,
		seeds:    seeds,
		acc:      account.NewAccount(""),
		host:     host,
		nodeType: nodeTypeSeed,
	}
}

func generateNormNode(gov []keypair.PublicKey, seeds []string, host string) *wrapNode {
	return &wrapNode{
		gov:      gov,
		seeds:    seeds,
		acc:      account.NewAccount(""),
		host:     host,
		nodeType: nodeTypeNorm,
	}
}

func generateGovNode(gov []keypair.PublicKey, seeds []string, host string, acc *account.Account) *wrapNode {
	return &wrapNode{
		gov:      gov,
		seeds:    seeds,
		acc:      acc,
		host:     host,
		nodeType: nodeTypeGov,
	}
}

func (wn *wrapNode) generateNode() error {
	node, err := GenerateNetServerWithSubnet(wn.gov, wn.acc, wn.seeds, wn.host)
	if err != nil {
		return err
	}
	wn.node = node
	return nil
}

func assembleIpAndPort(ip string, port uint16) string {
	return fmt.Sprintf("%s:%d", ip, port)
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

func settleDefConfigPort(addr string) error {
	_, port, err := net.SplitHostPort(addr)
	if err != nil {
		return err
	}
	iport, err := strconv.Atoi(port)
	if err != nil {
		return err
	}
	conf.DefConfig.Net.NodePort = uint16(iport)
	return nil
}
