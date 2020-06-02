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
	log4 "github.com/alecthomas/log4go"
	"github.com/ontology-community/onRobot/internal/robot/conf"
	"github.com/ontology-community/onRobot/pkg/p2pserver/net/netserver"
	"github.com/ontology-community/onRobot/pkg/p2pserver/net/protocol"
)

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
	return
}
