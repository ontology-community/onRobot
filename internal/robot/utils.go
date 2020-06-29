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
	"time"

	"github.com/ontio/ontology/account"
	"github.com/ontio/ontology/common/log"
	onthttp "github.com/ontio/ontology/http/base/common"
	"github.com/ontology-community/onRobot/internal/robot/conf"
	"github.com/ontology-community/onRobot/pkg/sdk"

	p2pcm "github.com/ontology-community/onRobot/pkg/p2pserver/common"
	"github.com/ontology-community/onRobot/pkg/p2pserver/message/types"
	"github.com/ontology-community/onRobot/pkg/p2pserver/net/netserver"
	"github.com/ontology-community/onRobot/pkg/p2pserver/net/protocol"
	"github.com/ontology-community/onRobot/pkg/p2pserver/params"
	"github.com/ontology-community/onRobot/pkg/p2pserver/peer"
	"github.com/ontology-community/onRobot/pkg/p2pserver/protocols"
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
	//staticFilter := connect_controller.NewStaticReserveFilter(nil)
	//reserved := protocol.GetReservedAddrFilter(false)
	//reservedPeers := p2p.CombineAddrFilter(staticFilter, reserved)

	if ns, err = netserver.NewNetServer(protocol, conf.DefConfig.Net, nil); err != nil {
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
	if ns, err = netserver.NewNetServer(protocol, conf.DefConfig.Net, nil); err != nil {
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
	if ns, err = netserver.NewNetServerWithKid(protocol, conf.DefConfig.Net, kid, nil); err != nil {
		log.Fatalf("[NewNetServer] crashed, err %s", err)
		return nil
	}
	if err = ns.Start(); err != nil {
		log.Fatalf("start netserver failed, err %s", err)
	}
	nsList = append(nsList, ns)
	return
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

// dispatch
func dispatch(sec int) {
	expire := time.Duration(sec) * time.Second
	time.Sleep(expire)
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
