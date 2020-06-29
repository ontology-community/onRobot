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
	"time"

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

type Reserve struct {
	Host string
	Rsv  []string
}

type ReserveList []*Reserve

func (r ReserveList) GetRsv(addr string) []string {
	if r == nil {
		return nil
	}

	for _, rsv := range r {
		host, _, err := net.SplitHostPort(addr)
		if err != nil {
			continue
		}
		if rsv.Host == host {
			return rsv.Rsv
		}
	}
	return nil
}

type MockSubnetConfig struct {
	Seeds, Govs, Norms []string
}

func (c *MockSubnetConfig) getLength() (S, G, N, T int) {
	S, G, N = len(c.Seeds), len(c.Govs), len(c.Norms)
	T = S + G + N
	return
}

func (c *MockSubnetConfig) combineNodes() []string {
	list := make([]string, 0)
	list = append(list, c.Seeds...)
	list = append(list, c.Govs...)
	list = append(list, c.Norms...)
	return list
}

func (c *MockSubnetConfig) checkDumpIps() error {
	exist := make(map[string]struct{})
	list := c.combineNodes()
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

func (c *MockSubnetConfig) IsSeedNode(addr string) bool {
	for _, host := range c.Seeds {
		if addr == host {
			return true
		}
	}
	return false
}

func (c *MockSubnetConfig) IsGovNode(addr string) bool {
	for _, host := range c.Govs {
		if addr == host {
			return true
		}
	}
	return false
}

func (c *MockSubnetConfig) CheckRsvs(rsvs ReserveList) error {
	nodes := c.combineNodes()
	for _, c := range rsvs {
		if !nodeInList(c.Host, nodes) {
			return fmt.Errorf("host %s not in subnet config", c.Host)
		}
		for _, rsv := range c.Rsv {
			if !nodeInList(rsv, nodes) {
				return fmt.Errorf("rsv %s not in subnet config", rsv)
			}
		}
	}
	return nil
}

type MockSubnet struct {
	c *MockSubnetConfig

	nodes  []*wrapNode
	net    mock.Network      // 共用同一个network
	ledger *utils.MockLedger // 共用同一个resolver 模拟从合约获取共识节点列表
}

func NewMockSubnet(c *MockSubnetConfig) (*MockSubnet, error) {
	return NewCustomMockSubnet(c, nil)
}

func NewMockSubnetWithReserves(c *MockSubnetConfig, rsvs ReserveList) (*MockSubnet, error) {
	return NewCustomMockSubnet(c, rsvs)
}

func NewCustomMockSubnet(c *MockSubnetConfig, rsvs ReserveList) (*MockSubnet, error) {
	if err := c.checkDumpIps(); err != nil {
		return nil, err
	}

	_, G, _, T := c.getLength()
	ms := &MockSubnet{
		c:     c,
		nodes: make([]*wrapNode, 0, T),
		net:   mock.NewNetwork(),
	}

	govPubKeys, gov := generateMultiPubkeys(G)
	ms.ledger = utils.NewMockLedger()
	for _, kp := range govPubKeys {
		ms.ledger.AddGovNode(kp)
	}

	for _, addr := range c.Seeds {
		rsv := rsvs.GetRsv(addr)
		wn := ms.generateNode(addr, nodeTypeSeed, nil, rsv)
		ms.nodes = append(ms.nodes, wn)
	}
	for i, addr := range c.Govs {
		rsv := rsvs.GetRsv(addr)
		wn := ms.generateNode(addr, nodeTypeGov, gov[i], rsv)
		ms.nodes = append(ms.nodes, wn)
	}
	for _, addr := range c.Norms {
		rsv := rsvs.GetRsv(addr)
		wn := ms.generateNode(addr, nodeTypeNorm, nil, rsv)
		ms.nodes = append(ms.nodes, wn)
	}
	return ms, nil
}

func (ms *MockSubnet) StartAll() {
	seeds := ms.nodes[:len(ms.c.Seeds)]
	others := ms.nodes[len(ms.c.Seeds):]
	for _, node := range seeds {
		go node.node.Start()
	}
	time.Sleep(3 * time.Second)
	for _, node := range others {
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
	wn := ms.generateNode(addr, nodeTypeGov, acc, nil)
	ms.nodes = append(ms.nodes, wn)
	return wn, nil
}

// 删除共识节点但是其本身并不关停，而是变成普通同步节点
func (ms *MockSubnet) DelGovNode(addr string) (wn *wrapNode, err error) {
	var acc keypair.PublicKey
	for _, node := range ms.nodes {
		if node.host == addr {
			wn = node
			if node.nodeType != nodeTypeGov {
				err = fmt.Errorf("node %s is not gov node", addr)
				return
			}
			node.nodeType = nodeTypeNorm
			acc = node.acc.PublicKey
		}
	}
	if wn == nil {
		err = fmt.Errorf("node %s not exist in gov node list", addr)
		return
	}
	for i, gov := range ms.c.Govs {
		if gov == addr {
			ms.c.Govs = append(ms.c.Govs[:i], ms.c.Govs[i+1:]...)
			ms.c.Norms = append(ms.c.Norms, addr)
		}
	}
	ms.ledger.DelGovNode(acc)
	return
}

/*
共识节点也是种子节点的情况，既担任共识节点的角色维持subnet网络列表，过滤掉subnet节点ip发往普通节点；也担任种子节点的角色，允许所有普通节点进行连接和区块同步。
*/
func (ms *MockSubnet) ReGenerateGovNodeInSeed(gov string) {
	origin := ms.c.Seeds
	for i, node := range ms.nodes {
		if node.host == gov && node.nodeType == nodeTypeGov {
			ms.c.Seeds = append(ms.c.Seeds, gov)
			wn := ms.generateNode(node.host, node.nodeType, node.acc, nil)
			ms.nodes[i] = wn
		}
	}
	ms.c.Seeds = origin
}

func (ms *MockSubnet) generateNode(host string, typ nodeType, acc *account.Account, rsv []string) *wrapNode {
	if acc == nil {
		acc = account.NewAccount("")
	}
	wn := &wrapNode{
		cfg:      ms.c,
		acc:      acc,
		host:     host,
		nodeType: typ,
	}

	// generate netserver
	protocol := protocols.NewSubnetHandler(acc, ms.c.Seeds, ms.ledger)
	resvFilter := protocol.GetReservedAddrFilter(len(rsv) != 0)
	wn.node = netserver.NewNetServerWithSubset(host, protocol, ms.net, rsv, resvFilter)

	return wn
}

func (ms *MockSubnet) CheckAll() error {
	for _, wn := range ms.nodes {
		log.Infof("===============================[check %s node %s]=================================",
			wn.typeName(), wn.host)

		if err := wn.checkMemberInfo(); err != nil {
			return err
		}
		if err := wn.checkNeighbors(); err != nil {
			return err
		}
	}
	log.Info("-----------------------------[end check]----------------------------------")
	return nil
}

func (ms *MockSubnet) CheckGovSeed(gov string) error {
	var wn *wrapNode
	for _, node := range ms.nodes {
		if node.host == gov {
			wn = node
		}
	}
	if wn == nil {
		return fmt.Errorf("govSeed node %s not exist", gov)
	}

	norms := make([]string, 0)
	nbs := wn.node.GetNeighbors()
	for _, nb := range nbs {
		addr := nb.GetAddr()
		if wn.isNormNodes(addr) {
			norms = append(norms, addr)
		}
	}

	if len(norms) == 0 {
		return fmt.Errorf("govSeed node %s has no normal node as neighbor", gov)
	}

	for _, norm := range norms {
		for _, node := range ms.nodes {
			if node.host == norm {
				mems := len(node.getSubnetMemberInfo())
				if mems > 0 {
					return fmt.Errorf("govSeed node %s, normal neighbor %s has %d subnet members",
						gov, norm, mems)
				}
			}
		}
	}

	return nil
}

func (ms *MockSubnet) CheckReserve() error {
	for _, wn := range ms.nodes {
		log.Infof("===============================[check %s node %s]=================================",
			wn.typeName(), wn.host)

		for _, member := range wn.getSubnetMemberInfo() {
			log.Infof("local %s, subnet members %s, connected %t", wn.host, member.ListenAddr, member.Connected)
		}

		for _, nb := range wn.node.GetNeighbors() {
			log.Infof("local %s, neighbor %s", wn.host, nb.GetAddr())
		}
	}
	return nil
}

//////////////////////////////////////////////////////
//
// wrapNode
//
//////////////////////////////////////////////////////
type wrapNode struct {
	cfg      *MockSubnetConfig
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
	if len(memList) != len(wn.cfg.Govs) {
		return fmt.Errorf("subnet member length err, member length( %d) should be %d", len(memList), len(wn.cfg.Govs))
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
	var num int

	S, G, N, _ := wn.cfg.getLength()
	switch wn.nodeType {
	case nodeTypeSeed:
		num = S + G + N - 1
	case nodeTypeGov:
		num = S + G - 1
	case nodeTypeNorm:
		num = S + N - 1
	}

	if L != num {
		return fmt.Errorf("%s node neighbor count(%d) shoud be %d", wn.typeName(), L, num)
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
	return nodeInList(addr, wn.cfg.Govs)
}

func (wn *wrapNode) isSeedNodes(addr string) bool {
	return nodeInList(addr, wn.cfg.Seeds)
}

func (wn *wrapNode) isNormNodes(addr string) bool {
	return nodeInList(addr, wn.cfg.Norms)
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
