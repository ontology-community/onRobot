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

package main

import (
	"flag"

	"github.com/ontio/ontology/common/log"
	"github.com/ontology-community/onRobot/internal/p2pnode"
	"github.com/ontology-community/onRobot/internal/p2pnode/conf"
)

var (
	Config string

	httpPort,
	nodePort uint

	loglevel int
	logDir   string
)

func init() {
	flag.StringVar(&Config, "config", "target/node/config.json", "Config of ontology-tool")
	flag.UintVar(&httpPort, "httpport", 30001, "http info port")
	flag.UintVar(&nodePort, "nodeport", 40001, "p2pnode port")
	flag.IntVar(&loglevel, "loglevel", 2, "loglevel [1: debug, 2: info]")
	flag.StringVar(&logDir, "logdir", "target/node/log", "set log dir")
	flag.Parse()
}

func main() {
	log.InitLog(loglevel, log.Stdout)
	if err := conf.DefConfig.Init(Config, nodePort, httpPort); err != nil {
		panic(err)
	}
	p2pnode.NewNode()
}
