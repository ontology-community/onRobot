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
	log4 "github.com/alecthomas/log4go"
	"github.com/ontio/ontology/account"
	ontcm "github.com/ontio/ontology/common"
	onthttp "github.com/ontio/ontology/http/base/common"
	"github.com/ontology-community/onRobot/internal/robot/conf"
	p2pcm "github.com/ontology-community/onRobot/pkg/p2pserver/common"
	"github.com/ontology-community/onRobot/pkg/p2pserver/httpinfo"
	"github.com/ontology-community/onRobot/pkg/p2pserver/message/types"
	"github.com/ontology-community/onRobot/pkg/p2pserver/net/netserver"
	"github.com/ontology-community/onRobot/pkg/p2pserver/net/protocol"
	"github.com/ontology-community/onRobot/pkg/p2pserver/params"
	"github.com/ontology-community/onRobot/pkg/p2pserver/peer"
	"github.com/ontology-community/onRobot/pkg/p2pserver/protocols"
	"github.com/ontology-community/onRobot/pkg/sdk"
	"io"
	"io/ioutil"
	"math"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

const (
	MaxNetServerNumber = 128
)

var (
	nsList = make([]*netserver.NetServer, 0, MaxNetServerNumber)
)

func reset() {
	log4.Debug("[GC] end testing, stop server and clear instance...")
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
		log4.Crashf("[NewNetServer] crashed, err %s", err)
	}
	if err = ns.Start(); err != nil {
		log4.Crashf("start netserver failed, err %s", err)
	}
	nsList = append(nsList, ns)
	return
}

// GenerateNetServerWithContinuePort
func GenerateNetServerWithContinuePort(protocol p2p.Protocol, port uint16) (ns *netserver.NetServer) {
	var err error

	conf.DefConfig.Net.NodePort = port
	if ns, err = netserver.NewNetServer(protocol, conf.DefConfig.Net); err != nil {
		log4.Crashf("[NewNetServer] crashed, err %s", err)
		return nil
	}
	if err = ns.Start(); err != nil {
		log4.Crashf("start netserver failed, err %s", err)
	}
	nsList = append(nsList, ns)
	return
}

func GenerateNetServerWithFakeIP(protocol p2p.Protocol, port uint16, mtx *sync.Mutex) (ns *netserver.NetServer) {
	var err error

	mtx.Lock()
	conf.DefConfig.Net.NodePort = port
	if ns, err = netserver.NewNetServer(protocol, conf.DefConfig.Net); err != nil {
		mtx.Unlock()
		log4.Crashf("[NewNetServer] crashed, err %s", err)
		return nil
	}
	mtx.Unlock()
	if err = ns.Start(); err != nil {
		log4.Crashf("start netserver failed, err %s", err)
	}
	nsList = append(nsList, ns)
	return
}

func GenerateNetServerWithFakeKid(protocol p2p.Protocol, kid *p2pcm.PeerKeyId) (ns *netserver.NetServer) {
	var err error
	if ns, err = netserver.NewNetServerWithKid(protocol, conf.DefConfig.Net, kid); err != nil {
		log4.Crashf("[NewNetServer] crashed, err %s", err)
		return nil
	}
	if err = ns.Start(); err != nil {
		log4.Crashf("start netserver failed, err %s", err)
	}
	nsList = append(nsList, ns)
	return
}

// GenerateMultiHeartbeatOnlyPeers
func GenerateMultiHeartbeatOnlyPeers(remoteList []string) ([]*peer.Peer, error) {
	protocol := protocols.NewOnlyHeartbeatMsgHandler()
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
		log4.Debug("current block height %d", curHeight)
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
	log4.Info("transaction hash %s", hash.ToHexString())
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
		log4.Info("transaction hash %s", hash.ToHexString())
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

// Dispatch
func Dispatch(sec int) {
	expire := time.Duration(sec) * time.Second
	time.Sleep(expire)
}

type statCount struct {
	send uint64
	recv uint64
	mu   *sync.Mutex
}

type httpClient struct {
	cli *http.Client
}

func NewHttpClient() *httpClient {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	return &httpClient{
		cli: client,
	}
}

func (c *httpClient) statMsgCount(iplist []string, startHttpPort, endHttpPort uint16, stat *statCount) {
	log4.Debug("init sendCount:%d, recvCount:%d", stat.send, stat.recv)

	for _, ip := range iplist {
		for p := startHttpPort; p <= endHttpPort; p++ {
			count, err := c.getStatResult(ip, p, "/stat/send")
			if err != nil {
				_ = log4.Error("[send/count] err:%s", err)
			} else {
				stat.mu.Lock()
				stat.send = count
				stat.mu.Unlock()
			}

			count, err = c.getStatResult(ip, p, "/stat/recv")
			if err != nil {
				_ = log4.Error("[recv/count] err:%s", err)
			} else {
				stat.mu.Lock()
				stat.recv = count
				stat.mu.Unlock()
			}
		}
	}
}

func (c *httpClient) clearMsgCount(iplist []string, startHttpPort, endHttpPort uint16) {
	log4.Debug("clear msg stat")

	for _, ip := range iplist {
		for p := startHttpPort; p <= endHttpPort; p++ {
			if err := c.getClearResult(ip, p, "/stat/clear"); err != nil {
				_ = log4.Error("[send/count] err:%s", err)
			}
		}
	}
}

func (c *httpClient) getStatResult(ip string, port uint16, method string) (count uint64, err error) {
	var (
		req  *http.Request
		resp *http.Response
		res  = &httpinfo.Resp{}
		num  float64
		ok   bool
	)

	url := fmt.Sprintf("http://%s:%d%s", ip, port, method)
	if req, err = http.NewRequest("GET", url, nil); err != nil {
		return
	}
	if resp, err = c.cli.Do(req); err != nil {
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
	if num, ok = res.Data.(float64); !ok {
		err = fmt.Errorf("stat count type invalid")
		return
	}
	count = uint64(num)
	return
}

func (c *httpClient) getClearResult(ip string, port uint16, method string) (err error) {
	var (
		req  *http.Request
		resp *http.Response
		res  = &httpinfo.Resp{}
	)

	url := fmt.Sprintf("http://%s:%d%s", ip, port, method)
	if req, err = http.NewRequest("GET", url, nil); err != nil {
		return
	}
	if resp, err = c.cli.Do(req); err != nil {
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
	dst  chan string
	stop chan struct{}
}

func NewInvalidTxWorker(remote string) (*invalidTxWorker, error) {
	protocol := protocols.NewOnlyHeartbeatMsgHandler()
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
		stop: make(chan struct{}),
		dst:  make(chan string),
	}, nil
}

func (w *invalidTxWorker) Start() {
	for {
		select {
		case dst := <-w.dst:
			if err := w.sendInvalidTxWithoutCheckBalance(dst); err != nil {
				_ = log4.Error("%s", err)
			}
		case <-w.stop:
			return
		}
	}
}

func (w *invalidTxWorker) Stop() {
	close(w.stop)
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
	err = w.pr.Send(tran)
	return
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

	protocol := protocols.NewOnlyHeartbeatMsgHandler()
	ns := GenerateNetServerWithProtocol(protocol)
	pr, err := ns.ConnectAndReturnPeer(remote)
	if err != nil {
		return err
	}
	if err := pr.Send(tran); err != nil {
		return err
	}

	Dispatch(expire)
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
		log4.Debug("===== node %s, origin tx %s, succeed tx %s", jsonrpc, hash.ToHexString(), hash1.ToHexString())
	} else {
		_ = log4.Error("===== node %s, origin tx %s failed", jsonrpc, hash.ToHexString())
	}

	log4.Info("===== src address %s, dst address %s", acc.Address.ToBase58(), dest)
	log4.Info("===== before transfer, src %s, dst %s, ", srcbfTx.Ont, dstbfTx.Ont)
	log4.Info("===== after transfer, src %s, dst %s, ", srcafTx.Ont, dstafTx.Ont)

	return nil
}
