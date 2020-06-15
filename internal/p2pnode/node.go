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

package p2pnode

import (
	"fmt"

	"github.com/ontio/ontology/account"
	"github.com/ontology-community/onRobot/pkg/dao"
	"github.com/ontology-community/onRobot/pkg/sdk"

	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/events"
	"github.com/ontology-community/onRobot/internal/p2pnode/conf"
	"github.com/ontology-community/onRobot/pkg/p2pserver"
	netreqactor "github.com/ontology-community/onRobot/pkg/p2pserver/actor/req"
	"github.com/ontology-community/onRobot/pkg/p2pserver/httpinfo"
	"github.com/ontology-community/onRobot/pkg/p2pserver/net/netserver"
	p2p "github.com/ontology-community/onRobot/pkg/p2pserver/net/protocol"
	"github.com/ontology-community/onRobot/pkg/p2pserver/protocols"
	"github.com/ontology-community/onRobot/pkg/txnpool"
	"github.com/ontology-community/onRobot/pkg/txnpool/proc"
)

var (
	acc *account.Account
)

func NewNode(walletpath, pwd string) {
	events.Init()
	initMysql()

	if err := initAccount(walletpath, pwd); err != nil {
		log.Fatal(err)
	}

	tp, err := initTxPool()
	if err != nil {
		log.Fatal(err)
	}
	msghandler := initProtocol()

	node, err := initP2PNode(tp, msghandler)
	if err != nil {
		log.Fatal(err)
	}

	ns := node.GetNetwork().(*netserver.NetServer)
	httpinfo.RunTxInfoHttpServer(ns, conf.DefConfig.Net.HttpInfoPort)
}

func initAccount(walletpath, pwd string) error {
	var err error
	if acc, err = sdk.RecoverAccount(walletpath, pwd); err != nil {
		return err
	}
	return nil
}

func initMysql() {
	dao.NewDao(conf.DefConfig.Mysql)
}

func initTxPool() (*proc.TXPoolServer, error) {

	txPoolServer, err := txnpool.StartTxnPoolServer()
	if err != nil {
		return nil, fmt.Errorf("Init txpool error: %s", err)
	}

	//hserver.SetTxPid(txPoolServer.GetPID())

	log.Info("TxPool init success")
	return txPoolServer, nil
}

func initProtocol() p2p.Protocol {
	return protocols.NewTxCountHandler(acc)
}

func initP2PNode(txpoolSvr *proc.TXPoolServer, handler p2p.Protocol) (*p2pserver.P2PServer, error) {
	p2p, err := p2pserver.NewStatServer(handler, conf.DefConfig.Net)
	if err != nil {
		return nil, err
	}
	if err := p2p.Start(); err != nil {
		return nil, fmt.Errorf("p2p service start error %s", err)
	}
	netreqactor.SetTxnPoolPid(txpoolSvr.GetPID())
	txpoolSvr.Net = p2p.GetNetwork()
	log.Info("P2P init success")
	return p2p, nil
}
