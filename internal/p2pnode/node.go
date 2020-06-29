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

	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/events"
	"github.com/ontology-community/onRobot/internal/p2pnode/conf"
	"github.com/ontology-community/onRobot/pkg/dao"
	"github.com/ontology-community/onRobot/pkg/p2pserver"
	netreqactor "github.com/ontology-community/onRobot/pkg/p2pserver/actor/req"
	"github.com/ontology-community/onRobot/pkg/p2pserver/httpinfo"
	"github.com/ontology-community/onRobot/pkg/p2pserver/net/netserver"
	"github.com/ontology-community/onRobot/pkg/p2pserver/protocols"
	"github.com/ontology-community/onRobot/pkg/txnpool"
)

func NewNode() {
	events.Init()

	dao.NewDao(conf.DefConfig.Mysql)
	msghandler := protocols.NewTxCountHandler()

	node, err := p2pserver.NewStatServer(msghandler, conf.DefConfig.Net)
	if err != nil {
		panic(err)
	}
	if err := node.Start(); err != nil {
		panic(err)
	}

	txpoolSvr, err := txnpool.StartTxnPoolServer()
	if err != nil {
		panic(fmt.Sprintf("Init txpool error: %s", err))
	}
	netreqactor.SetTxnPoolPid(txpoolSvr.GetPID())
	txpoolSvr.Net = node.GetNetwork()

	ns := node.GetNetwork().(*netserver.NetServer)
	httpinfo.RunTxInfoHttpServer(ns, conf.DefConfig.Net.HttpInfoPort)

	log.Info("P2P init success")
}
