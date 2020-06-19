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

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/account"
	"github.com/ontio/ontology/common/log"

	p2pcm "github.com/ontology-community/onRobot/pkg/p2pserver/common"
	"github.com/ontology-community/onRobot/pkg/p2pserver/mock"
	"github.com/ontology-community/onRobot/pkg/p2pserver/net/netserver"
	"github.com/ontology-community/onRobot/pkg/p2pserver/protocols"
	"github.com/ontology-community/onRobot/pkg/p2pserver/protocols/utils"
)

type nodeType uint

const (
	nodeTypeUnknown nodeType = iota
	nodeTypeSeed
	nodeTypeGov
	nodeTypeNorm
)

type MockSubnetConfig struct {
	Seeds, Govs, Norms []string
}

func (c *MockSubnetConfig) getLength() (S, G, N, T int) {
	S, G, N = len(c.Seeds), len(c.Govs), len(c.Norms)
	T = S + G + N
	return
}

func (c *MockSubnetConfig) checkDumpIps() error {
	exist := make(map[string]struct{})
	list := make([]string, 0)
	list = append(list, c.Seeds...)
	list = append(list, c.Govs...)
	list = append(list, c.Norms...)
	for _, addr := range list {
		host, _, _ := net.SplitHostPort(addr)
		if _, ok := exist[host]; ok {
			return fmt.Errorf("ip %s already exist", host)
		} else {
			exist[host] = struct{}{}
		}
	}
	return nil
}

type MockSubnet struct {
	c *MockSubnetConfig

	nodes  []*wrapNode
	nw     mock.Network      // 共用同一个network
	ledger *utils.MockLedger // 共用同一个resolver 模拟从合约获取共识节点列表
}

func NewMockSubnet(c *MockSubnetConfig) (*MockSubnet, error) {
	if err := c.checkDumpIps(); err != nil {
		return nil, err
	}

	_, G, _, T := c.getLength()
	ms := &MockSubnet{
		c:     c,
		nodes: make([]*wrapNode, 0, T),
		nw:    mock.NewNetwork(),
	}

	govPubKeys, govAccounts := generateMultiPubkeys(G)
	ms.ledger = utils.NewMockLedger()
	for _, kp := range govPubKeys {
		ms.ledger.AddGovNode(kp)
	}

	for _, addr := range c.Seeds {
		ms.generateNode(addr, nodeTypeSeed, nil)
	}
	for i, addr := range c.Govs {
		ms.generateNode(addr, nodeTypeGov, govAccounts[i])
	}
	for _, addr := range c.Norms {
		ms.generateNode(addr, nodeTypeNorm, nil)
	}

	return ms, nil
}

func (ms *MockSubnet) StartAll() {
	for _, node := range ms.nodes {
		go node.node.Start()
	}
}

func (ms *MockSubnet) AddGovNode(addr string) (*wrapNode, error) {
	ms.c.Govs = append(ms.c.Govs, addr)
	if err := ms.c.checkDumpIps(); err != nil {
		return nil, err
	}

	pubkey, acc := generateSinglePubkey()
	ms.ledger.AddGovNode(pubkey)
	wn := ms.generateNode(addr, nodeTypeGov, acc)
	return wn, nil
}

func (ms *MockSubnet) DelGovNode(addr string) {

}

func (ms *MockSubnet) CheckAll() error {
	for _, node := range ms.nodes {
		log.Infof("===============================[check %s node %s]=================================",
			node.typeName(), node.host)

		if err := node.checkMemberInfo(); err != nil {
			return err
		}
		if err := node.checkNeighbors(); err != nil {
			return err
		}
	}
	log.Info("-----------------------------[end check]----------------------------------")
	return nil
}

func (ms *MockSubnet) generateNode(host string, typ nodeType, acc *account.Account) *wrapNode {
	if acc == nil {
		acc = account.NewAccount("")
	}
	wn := &wrapNode{
		c:        ms.c,
		acc:      acc,
		host:     host,
		nodeType: typ,
	}

	// generate netserver
	protocol := protocols.NewSubnetHandler(acc, ms.c.Seeds, ms.ledger)
	wn.node = netserver.NewNetServerWithSubset(host, protocol, ms.nw)

	ms.nodes = append(ms.nodes, wn)
	return wn
}

// wrapNode
type wrapNode struct {
	c        *MockSubnetConfig
	node     *netserver.NetServer
	acc      *account.Account
	host     string
	nodeType nodeType
}

func (wn *wrapNode) checkMemberInfo() error {
	// 1. get subnet members
	memList := wn.getSubnetMemberInfo()

	// 2. normal node won't connect with any gov nodes
	if wn.nodeType == nodeTypeNorm {
		if len(memList) > 0 {
			return fmt.Errorf("norm node connected with %d gov nodes", len(memList))
		}
		return nil
	}

	// 3. check gov member and connecting situation
	for _, member := range memList {
		if !wn.isGovNodes(member.ListenAddr) {
			return fmt.Errorf("memeber %s is not gov node", member.ListenAddr)
		}
		if member.Connected == false {
			return fmt.Errorf("member %s connected broken", member.ListenAddr)
		}
		log.Infof("local %s, subnet member %s, connected %t", wn.host, member.ListenAddr, member.Connected)
	}

	// 4. check gov member length
	if len(memList) != len(wn.c.Govs) {
		return fmt.Errorf("gov length %d != mems lenth %d", len(memList), len(wn.c.Govs))
	}

	return nil
}

func (wn *wrapNode) checkNeighbors() error {
	nbs := wn.node.GetNeighbors()
	for _, nb := range nbs {
		addr := nb.GetAddr()
		if wn.isSelf(addr) {
			return fmt.Errorf("node %s connected itself")
		}
		if nodeTyp, err := wn.checkNeighborNode(addr); err != nil {
			return err
		} else {
			log.Infof("local %s, neighbor %s, nodetype %s", wn.host, addr, nodeTyp)
		}
	}

	return wn.checkNeighborCount(len(nbs))
}

func (wn *wrapNode) checkNeighborNode(addr string) (typName string, err error) {
	remoteTyp := wn.remoteNodeType(addr)

	switch wn.nodeType {
	case nodeTypeSeed:
	case nodeTypeGov:
		if remoteTyp == nodeTypeNorm {
			err = fmt.Errorf("gov node %s should not be connected by normal node %s", wn.host, addr)
		}
	case nodeTypeNorm:
		if remoteTyp == nodeTypeGov {
			err = fmt.Errorf("norm node %s should not be connected by gov node %s", wn.host, addr)
		}
	default:
		err = fmt.Errorf("invalid node type")
	}
	typName = type2Name(remoteTyp)
	return
}

func (wn *wrapNode) checkNeighborCount(L int) error {
	S, G, N, _ := wn.c.getLength()
	switch wn.nodeType {
	case nodeTypeSeed:
		if L != S+G+N-1 {
			return fmt.Errorf("seed node neighbor count %d invalid", L)
		}
	case nodeTypeGov:
		if L != S+G-1 {
			return fmt.Errorf("gov node neighbor count %d invalid", L)
		}
	case nodeTypeNorm:
		if L != S+N-1 {
			return fmt.Errorf("norm node neighbor count %d invalid", L)
		}
	}
	return nil
}

func (wn *wrapNode) getSubnetMemberInfo() []p2pcm.SubnetMemberInfo {
	protocol := wn.node.Protocol()
	handler, ok := protocol.(*protocols.SubnetHandler)
	if !ok {
		return nil
	}

	return handler.GetSubnetMembersInfo()
}

func (wn *wrapNode) remoteNodeType(addr string) nodeType {
	if wn.isSeedNodes(addr) {
		return nodeTypeSeed
	} else if wn.isGovNodes(addr) {
		return nodeTypeGov
	} else if wn.isNormNodes(addr) {
		return nodeTypeNorm
	} else {
		return nodeTypeUnknown
	}
}

func (wn *wrapNode) isGovNodes(addr string) bool {
	return nodeInList(addr, wn.c.Govs)
}

func (wn *wrapNode) isSeedNodes(addr string) bool {
	return nodeInList(addr, wn.c.Seeds)
}

func (wn *wrapNode) isNormNodes(addr string) bool {
	return nodeInList(addr, wn.c.Norms)
}

func (wn *wrapNode) isSelf(addr string) bool {
	host, _, _ := net.SplitHostPort(addr)
	listenHost, _, _ := net.SplitHostPort(wn.host)
	return host == listenHost
}

func (wn *wrapNode) typeName() string {
	return type2Name(wn.nodeType)
}

func type2Name(typ nodeType) string {
	var name string

	switch typ {
	case nodeTypeSeed:
		name = "seed"
	case nodeTypeGov:
		name = "gov"
	case nodeTypeNorm:
		name = "norm"
	default:
		name = "unknown"
	}
	return name
}

func nodeInList(addr string, list []string) bool {
	host, _, _ := net.SplitHostPort(addr)
	for _, listenAddr := range list {
		listenHost, _, _ := net.SplitHostPort(listenAddr)
		if host == listenHost {
			return true
		}
	}
	return false
}

func generateMultiPubkeys(n int) ([]keypair.PublicKey, []*account.Account) {
	accList := make([]*account.Account, n)
	pubkeyList := make([]keypair.PublicKey, n)
	for i := 0; i < n; i++ {
		pubkeyList[i], accList[i] = generateSinglePubkey()
	}
	return pubkeyList, accList
}

func generateSinglePubkey() (keypair.PublicKey, *account.Account) {
	acc := account.NewAccount("")
	return acc.PublicKey, acc
}
