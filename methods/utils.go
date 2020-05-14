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

package methods

import (
	"encoding/json"
	"fmt"
	log4 "github.com/alecthomas/log4go"
	"github.com/ontio/ontology/account"
	common2 "github.com/ontio/ontology/http/base/common"
	"github.com/ontology-community/onRobot/common"
	"github.com/ontology-community/onRobot/config"
	"github.com/ontology-community/onRobot/p2pserver/message/types"
	"github.com/ontology-community/onRobot/p2pserver/net/netserver"
	"github.com/ontology-community/onRobot/p2pserver/net/protocol"
	"github.com/ontology-community/onRobot/p2pserver/peer"
	"github.com/ontology-community/onRobot/p2pserver/protocols"
	"io/ioutil"
	"math/big"
	"math/rand"
	"os"
	"sync"
	"time"
)

// GenerateNetServerWithProtocol get netserver with some protocol
func GenerateNetServerWithProtocol(protocol p2p.Protocol) (ns *netserver.NetServer) {
	var err error

	if ns, err = netserver.NewNetServer(protocol, config.DefConfig.Net); err != nil {
		log4.Crashf("[NewNetServer] crashed, err %s", err)
	}
	if err = ns.Start(); err != nil {
		log4.Crashf("start netserver failed, err %s", err)
	}
	nsList = append(nsList, ns)
	return
}

// GenerateNetServerWithContinuePort
func GenerateNetServerWithContinuePort(protocol p2p.Protocol, port uint16, mtx *sync.Mutex) (ns *netserver.NetServer) {
	var err error

	mtx.Lock()
	config.DefConfig.Net.NodePort = port
	if ns, err = netserver.NewNetServer(protocol, config.DefConfig.Net); err != nil {
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

func GenerateNetServerWithFakeIP(protocol p2p.Protocol, port uint16, mtx *sync.Mutex) (ns *netserver.NetServer) {
	var err error

	mtx.Lock()
	config.DefConfig.Net.NodePort = port
	if ns, err = netserver.NewNetServer(protocol, config.DefConfig.Net); err != nil {
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

// GetAndSetBlockHeight get block height from other p2pserver and settle self height
func GetAndSetBlockHeight(jsonrpcAddr string, alpha uint64) (uint64, error) {
	curHeight, err := common.GetBlockCurrentHeight(jsonrpcAddr)
	if err != nil {
		return 0, err
	} else {
		log4.Debug("current block height %d", curHeight)
	}
	common.SetHeartbeatTestBlockHeight(curHeight + alpha)
	return curHeight, nil
}

// GetParamsFromJsonFile
func GetParamsFromJsonFile(fileName string, data interface{}) error {
	fullPath := config.ParamsFileDir + string(os.PathSeparator) + fileName
	bz, err := ioutil.ReadFile(fullPath)
	if err != nil {
		return err
	}
	return json.Unmarshal(bz, data)
}

// SettleBalanceListAndCompare get balance, settle in list and compare
func SettleBalanceListAndCompare(balanceList []*common2.BalanceOfRsp, jsonrpcList []string, acc *account.Account) error {
	if cap(balanceList) != len(jsonrpcList) {
		return fmt.Errorf("cap(balanceList) != len(jsonrpcList)")
	}

	for _, jsonRpc := range jsonrpcList {
		num, err := common.GetBalance(jsonRpc, acc.Address)
		if err != nil {
			return err
		}
		balanceList = append(balanceList, num)

		log4.Info("remote %s, %s balance before transfer %s",
			jsonRpc, acc.Address.ToBase58(), num.Ont)
	}

	cmp := balanceList[0]
	for _, balance := range balanceList[1:] {
		if cmp.Ont != balance.Ont {
			return fmt.Errorf("balance before transfer different")
		}
	}
	return nil
}

// GenerateTransferOntTx
func GenerateTransferOntTx(acc *account.Account, dst string, amount uint64) (*types.Trn, error) {
	addr, err := common.AddressFromBase58(dst)
	if err != nil {
		return nil, err
	}
	price := config.DefConfig.Sdk.GasPrice
	gas := config.DefConfig.Sdk.GasLimit
	tran, err := common.TransferOntTx(price, gas, acc, addr, amount)
	if err != nil {
		return nil, err
	}
	hash := tran.Hash()
	log4.Info("transaction hash %s", hash.ToHexString())
	tx := &types.Trn{Txn: tran}

	return tx, nil
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
	return peers, nil
}

// GenerateMultiRandomOntTransfer
func GenerateMultiRandomOntTransfer(acc *account.Account, dst string, initBalanceStr string, cap int) ([]*types.Trn, error) {
	list := make([]*types.Trn, 0, cap)

	outBoundBalance := int64(100000000)
	initBalance, _ := new(big.Int).SetString(initBalanceStr, 10)
	delta := rand.Int63n(outBoundBalance)
	amount := initBalance.Uint64() + uint64(delta)

	for i := 0; i < cap; i++ {
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

// Dispatch
func Dispatch(sec int) {
	expire := time.Duration(sec) * time.Second
	stop := make(chan struct{})
	tr.Add(expire, func() {
		stop <- struct{}{}
	})
	<-stop
}
