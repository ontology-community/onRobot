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
	log4 "github.com/alecthomas/log4go"
	"github.com/ontology-community/onRobot/config"
	"github.com/ontology-community/onRobot/p2pserver/net/netserver"
	"github.com/ontology-community/onRobot/p2pserver/net/protocol"
	"io/ioutil"
	"os"
	"sync"
	"time"
)

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

var paramsFileDir string

func SetParamsDir(path string) {
	paramsFileDir = path
}

func getParamsFromJsonFile(fileName string, data interface{}) error {
	fullPath := paramsFileDir + string(os.PathSeparator) + fileName
	bz, err := ioutil.ReadFile(fullPath)
	if err != nil {
		return err
	}
	return json.Unmarshal(bz, data)
}

func dispatch(sec int) {
	expire := time.Duration(sec) * time.Second
	stop := make(chan struct{})
	tr.Add(expire, func() {
		stop <- struct{}{}
	})
	<-stop
}
