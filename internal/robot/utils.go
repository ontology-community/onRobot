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
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"math/rand"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ontio/ontology-crypto/keypair"

	"github.com/ontio/ontology/account"
	ontcm "github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/log"
	onthttp "github.com/ontio/ontology/http/base/common"
	"github.com/ontology-community/onRobot/internal/robot/conf"
	"github.com/ontology-community/onRobot/pkg/sdk"

	p2pcm "github.com/ontology-community/onRobot/pkg/p2pserver/common"
	"github.com/ontology-community/onRobot/pkg/p2pserver/httpinfo"
	"github.com/ontology-community/onRobot/pkg/p2pserver/message/types"
	"github.com/ontology-community/onRobot/pkg/p2pserver/net/netserver"
	p2p "github.com/ontology-community/onRobot/pkg/p2pserver/net/protocol"
	"github.com/ontology-community/onRobot/pkg/p2pserver/params"
	"github.com/ontology-community/onRobot/pkg/p2pserver/peer"
	"github.com/ontology-community/onRobot/pkg/p2pserver/protocols"
	"github.com/ontology-community/onRobot/pkg/p2pserver/protocols/utils"
	"github.com/ontology-community/onRobot/pkg/p2pserver/stat"
)

const (
	MaxNetServerNumber = 128
)

var (
	nsList = make([]*netserver.NetServer, 0, MaxNetServerNumber)
)

func reset() {
	log.Debug("[GC] end testing, stop server and clear instance...")
	params.Reset()
	for _, ns := range nsList {
		if ns != nil {
			ns.Stop()
		}
	}
	nsList = make([]*netserver.NetServer, 0, MaxNetServerNumber)
}

// GenerateNetServerWithProtocol get netserver with some protocol
func GenerateNetServerWithProtocol(protocol p2p.Protocol) (ns *netserver.NetServer) {
	var err error

	if ns, err = netserver.NewNetServer(protocol, conf.DefConfig.Net); err != nil {
		log.Fatalf("[NewNetServer] crashed, err %s", err)
	}
	if err = ns.Start(); err != nil {
		log.Fatal("start netserver failed, err %s", err)
	}
	nsList = append(nsList, ns)
	return
}

// GenerateNetServerWithContinuePort
func GenerateNetServerWithContinuePort(protocol p2p.Protocol, port uint16) (ns *netserver.NetServer) {
	var err error

	conf.DefConfig.Net.NodePort = port
	if ns, err = netserver.NewNetServer(protocol, conf.DefConfig.Net); err != nil {
		log.Fatalf("[NewNetServer] crashed, err %s", err)
		return nil
	}
	if err = ns.Start(); err != nil {
		log.Fatalf("start netserver failed, err %s", err)
	}
	nsList = append(nsList, ns)
	return
}

func GenerateNetServerWithFakeKid(protocol p2p.Protocol, kid *p2pcm.PeerKeyId) (ns *netserver.NetServer) {
	var err error
	if ns, err = netserver.NewNetServerWithKid(protocol, conf.DefConfig.Net, kid); err != nil {
		log.Fatalf("[NewNetServer] crashed, err %s", err)
		return nil
	}
	if err = ns.Start(); err != nil {
		log.Fatalf("start netserver failed, err %s", err)
	}
	nsList = append(nsList, ns)
	return
}

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

// GenerateMultiHeartbeatOnlyPeers
func GenerateMultiHeartbeatOnlyPeers(remoteList []string) ([]*peer.Peer, error) {
	protocol := protocols.NewHeartbeatHandler()
	ns := GenerateNetServerWithProtocol(protocol)
	peers := make([]*peer.Peer, 0, len(remoteList))
	for _, remote := range remoteList {
		pr, err := ns.ConnectAndReturnPeer(remote)
		if err != nil {
			return nil, err
		}
		peers = append(peers, pr)
	}
	nsList = append(nsList, ns)
	return peers, nil
}

// GetAndSetBlockHeight get block height from other p2pserver and settle self height
func GetAndSetBlockHeight(jsonrpcAddr string, alpha uint64) (uint64, error) {
	curHeight, err := sdk.GetBlockCurrentHeight(jsonrpcAddr)
	if err != nil {
		return 0, err
	} else {
		log.Debugf("current block height %d", curHeight)
	}
	params.SetHeartbeatTestBlockHeight(curHeight + alpha)
	return curHeight, nil
}

// GetBalanceAndCompare get balance, settle in list and compare
func GetBalanceAndCompare(jsonrpcList []string, acc *account.Account) ([]*onthttp.BalanceOfRsp, error) {
	balanceList := make([]*onthttp.BalanceOfRsp, 0, len(jsonrpcList))
	for _, jsonRpc := range jsonrpcList {
		num, err := sdk.GetBalance(jsonRpc, acc.Address)
		if err != nil {
			return nil, err
		}
		balanceList = append(balanceList, num)
	}

	cmp := balanceList[0]
	for _, balance := range balanceList[1:] {
		if cmp.Ont != balance.Ont {
			return nil, fmt.Errorf("balance before transfer different")
		}
	}
	return balanceList, nil
}

func GetBlockHeightList(jsonrpcList []string) ([]uint64, error) {
	list := make([]uint64, 0, len(jsonrpcList))
	for _, jsonrpc := range jsonrpcList {
		height, err := sdk.GetBlockCurrentHeight(jsonrpc)
		if err != nil {
			return nil, err
		}
		list = append(list, height)
	}
	return list, nil
}

// GenerateTransferOntTx
func GenerateTransferOntTx(acc *account.Account, dst string, amount uint64) (*types.Trn, error) {
	addr, err := sdk.AddressFromBase58(dst)
	if err != nil {
		return nil, err
	}
	price := conf.DefConfig.Sdk.GasPrice
	gas := conf.DefConfig.Sdk.GasLimit
	tran, err := sdk.TransferOntTx(price, gas, acc, addr, amount)
	if err != nil {
		return nil, err
	}
	hash := tran.Hash()
	log.Infof("transaction hash %s", hash.ToHexString())
	tx := &types.Trn{Txn: tran}

	return tx, nil
}

// GenerateMultiOntTransfer
func GenerateMultiOntTransfer(acc *account.Account, dst string, amount uint64, num int) ([]*types.Trn, error) {
	list := make([]*types.Trn, 0, num)

	for i := 0; i < num; i++ {
		tran, err := GenerateTransferOntTx(acc, dst, amount)
		if err != nil {
			return nil, err
		}
		hash := tran.Txn.Hash()
		log.Infof("transaction hash %s", hash.ToHexString())
		list = append(list, tran)
	}

	return list, nil
}

// GenerateZeroDistancePeerIDs 生成距离为0的peerID列表
func GenerateZeroDistancePeerIDs(tgID p2pcm.PeerId, num int) ([]p2pcm.PeerId, error) {
	if num >= 128 {
		return nil, fmt.Errorf("list length should < 128")
	}

	list := make([]p2pcm.PeerId, 0, num)
	exists := make(map[uint64]struct{})

	sink := new(ontcm.ZeroCopySink)
	tgID.Serialization(sink)
	exists[tgID.ToUint64()] = struct{}{}

	var getValidXorByte = func(tg uint8) uint8 {
		var xor uint8
		for {
			delta := uint8(rand.Int63n(255))
			xor = delta ^ tg
			if xor >= 128 && xor <= 255 {
				break
			}
		}
		return xor
	}

	sinkbz := sink.Bytes()
	for {
		bz := new([20]byte)
		copy(bz[:], sinkbz[:])

		xor := getValidXorByte(sinkbz[0])
		bz[0] = xor

		source := ontcm.NewZeroCopySource(bz[:])
		peerID := p2pcm.PeerId{}
		if err := peerID.Deserialization(source); err != nil {
			continue
		}
		if _, exist := exists[peerID.ToUint64()]; exist {
			continue
		} else {
			exists[peerID.ToUint64()] = struct{}{}
			list = append(list, peerID)
			distance(peerID, tgID)
		}
		if len(list) >= num {
			break
		}
	}

	return list, nil
}

func distance(local, target p2pcm.PeerId) int {
	return p2pcm.CommonPrefixLen(local, target)
}

// dispatch
func dispatch(sec int) {
	expire := time.Duration(sec) * time.Second
	time.Sleep(expire)
}

type httpClient struct {
	ips []string
	startHttpPort,
	endHttpPort uint16

	list map[string]*http.Client
}

func NewHttpClient(ipList []string, startHttpPort, endHttpPort uint16) *httpClient {
	c := &httpClient{}
	c.ips = ipList
	c.startHttpPort = startHttpPort
	c.endHttpPort = endHttpPort
	c.list = make(map[string]*http.Client)

	for _, ip := range c.ips {
		for p := c.startHttpPort; p <= c.endHttpPort; p++ {
			c.setCli(ip, p)
		}
	}
	return c
}

func (c *httpClient) setCli(ip string, port uint16) {
	addr := assembleIpAndPort(ip, port)
	client := new(http.Client)
	client.Timeout = 10 * time.Second
	c.list[addr] = client
}

func (c *httpClient) getCli(ip string, port uint16) (*http.Client, error) {
	addr := assembleIpAndPort(ip, port)
	if cli, ok := c.list[addr]; !ok {
		return nil, fmt.Errorf("http client for %s:%d not exist", ip, port)
	} else {
		return cli, nil
	}
}

func (c *httpClient) statMsgCount() map[string]*stat.TxNum {
	list := make(map[string]*stat.TxNum)

	for _, ip := range c.ips {
		for p := c.startHttpPort; p <= c.endHttpPort; p++ {
			time.Sleep(100 * time.Millisecond)
			data, err := c.getStatResult(ip, p, httpinfo.StatList)
			if err != nil {
				log.Errorf("[send/count] err:%s", err)
			} else {
				for _, v := range data {
					if _, ok := list[v.Hash]; !ok {
						list[v.Hash] = &stat.TxNum{Hash: v.Hash, Send: 0, Recv: 0}
					}
					list[v.Hash].Send += v.Send
					list[v.Hash].Recv += v.Recv
				}
			}
		}
	}

	return list
}

func (c *httpClient) clearMsgCount() {
	log.Debug("clear msg stat")

	for _, ip := range c.ips {
		for p := c.startHttpPort; p <= c.endHttpPort; p++ {
			time.Sleep(100 * time.Millisecond)
			if err := c.getClearResult(ip, p, httpinfo.StatClear); err != nil {
				log.Errorf("[send/count] err:%s", err)
			}
		}
	}
}

type WrapTxNum struct {
	stat.TxNum
	Ip   string
	Port uint16
}

func (c *httpClient) statHashList(hashlist []string) []*WrapTxNum {
	list := make([]*WrapTxNum, 0)

	for _, ip := range c.ips {
		for p := c.startHttpPort; p <= c.endHttpPort; p++ {
			time.Sleep(100 * time.Millisecond)
			data, err := c.getHashListResult(ip, p, httpinfo.StatHashList, hashlist)
			if err != nil {
				log.Errorf("[send/count] err:%s", err)
			} else {
				for _, v := range data {
					tx := &WrapTxNum{Ip: ip, Port: p, TxNum: stat.TxNum{
						Hash: v.Hash, Send: v.Send, Recv: v.Recv,
					}}
					list = append(list, tx)
				}
			}
		}
	}

	return list
}

func (c *httpClient) getStatResult(ip string, port uint16, method string) (count []*stat.TxNum, err error) {
	var (
		req  *http.Request
		resp *http.Response
		res  = &httpinfo.Resp{}
		data string
		cli  *http.Client
		ok   bool
	)

	url := fmt.Sprintf("http://%s:%d%s", ip, port, method)
	if req, err = http.NewRequest("GET", url, nil); err != nil {
		return
	}
	if cli, err = c.getCli(ip, port); err != nil {
		return
	}
	if resp, err = cli.Do(req); err != nil {
		return
	}
	defer resp.Body.Close()
	if err = parseResponse(resp.Body, res); err != nil {
		return
	}
	if res.Succeed == false {
		err = fmt.Errorf("%s", res.Err)
		return
	}
	if data, ok = res.Data.(string); !ok {
		err = fmt.Errorf("stat count type invalid")
	}
	count, err = stat.Parse2TxNumList(data)
	return
}

func (c *httpClient) getClearResult(ip string, port uint16, method string) (err error) {
	var (
		req  *http.Request
		resp *http.Response
		res  = &httpinfo.Resp{}
		cli  *http.Client
	)

	url := fmt.Sprintf("http://%s:%d%s", ip, port, method)
	if req, err = http.NewRequest("GET", url, nil); err != nil {
		return
	}
	if cli, err = c.getCli(ip, port); err != nil {
		return
	}
	if resp, err = cli.Do(req); err != nil {
		return
	}
	defer resp.Body.Close()
	if err = parseResponse(resp.Body, res); err != nil {
		return
	}

	if res.Succeed == false {
		err = fmt.Errorf("%s", res.Err)
		return
	}

	return
}

func (c *httpClient) getHashListResult(ip string, port uint16, method string, hashlist []string) (count []*stat.TxNum, err error) {
	var (
		req  *http.Request
		resp *http.Response
		res  = &httpinfo.Resp{}
		data string
		cli  *http.Client
		ok   bool
	)

	reqparam := strings.Join(hashlist, ",")
	url := fmt.Sprintf("http://%s:%d%s?list=%s", ip, port, method, reqparam)
	if req, err = http.NewRequest("GET", url, nil); err != nil {
		return
	}
	if cli, err = c.getCli(ip, port); err != nil {
		return
	}
	if resp, err = cli.Do(req); err != nil {
		return
	}
	defer resp.Body.Close()
	if err = parseResponse(resp.Body, res); err != nil {
		return
	}
	if res.Succeed == false {
		err = fmt.Errorf("%s", res.Err)
		return
	}
	if data, ok = res.Data.(string); !ok {
		err = fmt.Errorf("stat count type invalid")
	}
	count, err = stat.Parse2TxNumList(data)
	return
}

func parseResponse(body io.Reader, res interface{}) error {
	data, err := ioutil.ReadAll(body)
	if err != nil {
		return fmt.Errorf("read http body error:%s", err)
	}

	err = json.Unmarshal(data, res)
	if err != nil {
		return fmt.Errorf("json.Unmarshal RestfulResp:%s error:%s", body, err)
	}

	return nil
}

type invalidTxWorker struct {
	pr   *peer.Peer
	acc  *account.Account
	list map[string]struct{}
	mu   *sync.RWMutex
}

func NewInvalidTxWorker(remote string) (*invalidTxWorker, error) {
	protocol := protocols.NewHeartbeatHandler()
	ns := GenerateNetServerWithProtocol(protocol)
	pr, err := ns.ConnectAndReturnPeer(remote)
	if err != nil {
		return nil, err
	}
	acc, err := sdk.RecoverAccount(conf.WalletPath, conf.DefConfig.WalletPwd)
	if err != nil {
		return nil, err
	}
	return &invalidTxWorker{
		pr:   pr,
		acc:  acc,
		list: make(map[string]struct{}),
		mu:   new(sync.RWMutex),
	}, nil
}

func (w *invalidTxWorker) sendMultiInvalidTx(num int, dest string) {
	for i := 0; i < num; i++ {
		go w.sendInvalidTxWithoutCheckBalance(dest)
	}
}

func (w *invalidTxWorker) sendInvalidTxWithoutCheckBalance(dest string) (err error) {
	var (
		tran   *types.Trn
		amount uint64 = math.MaxUint64
	)

	if tran, err = GenerateTransferOntTx(w.acc, dest, amount); err != nil {
		return
	}

	// send dump tx
	if err = w.pr.Send(tran); err != nil {
		return
	}

	// save in list
	hash := tran.Txn.Hash()
	w.mu.Lock()
	w.list[hash.ToHexString()] = struct{}{}
	w.mu.Unlock()

	return
}

func (w *invalidTxWorker) getHashList(length int) []string {
	w.mu.Lock()
	defer w.mu.Unlock()
	list := make([]string, 0)
	for v, _ := range w.list {
		list = append(list, v)
		delete(w.list, v)
		if len(list) >= length {
			break
		}
	}
	return list
}

func (w *invalidTxWorker) getMsgCount() int64 {
	w.mu.RLock()
	defer w.mu.RUnlock()
	num := len(w.list)
	return int64(num)
}

func singleTransfer(remote, jsonrpc, dest string, amount uint64, expire int) error {
	acc, err := sdk.RecoverAccount(conf.WalletPath, conf.DefConfig.WalletPwd)
	if err != nil {
		return err
	}
	destAddr, err := sdk.AddressFromBase58(dest)
	if err != nil {
		return err
	}
	srcbfTx, err := sdk.GetBalance(jsonrpc, acc.Address)
	if err != nil {
		return err
	}
	dstbfTx, err := sdk.GetBalance(jsonrpc, destAddr)
	if err != nil {
		return err
	}

	tran, err := GenerateTransferOntTx(acc, dest, amount)
	if err != nil {
		return err
	}
	hash := tran.Txn.Hash()

	protocol := protocols.NewHeartbeatHandler()
	ns := GenerateNetServerWithProtocol(protocol)
	pr, err := ns.ConnectAndReturnPeer(remote)
	if err != nil {
		return err
	}
	if err := pr.Send(tran); err != nil {
		return err
	}

	dispatch(expire)
	srcafTx, err := sdk.GetBalance(jsonrpc, acc.Address)
	if err != nil {
		return err
	}
	dstafTx, err := sdk.GetBalance(jsonrpc, destAddr)
	if err != nil {
		return err
	}

	tx, err := sdk.GetTxByHash(jsonrpc, hash)
	if err == nil {
		hash1 := tx.Hash()
		log.Debugf("===== node %s, origin tx %s, succeed tx %s", jsonrpc, hash.ToHexString(), hash1.ToHexString())
	} else {
		log.Errorf("===== node %s, origin tx %s failed", jsonrpc, hash.ToHexString())
	}

	log.Infof("===== src address %s, dst address %s", acc.Address.ToBase58(), dest)
	log.Infof("===== before transfer, src %s, dst %s, ", srcbfTx.Ont, dstbfTx.Ont)
	log.Infof("===== after transfer, src %s, dst %s, ", srcafTx.Ont, dstafTx.Ont)

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
