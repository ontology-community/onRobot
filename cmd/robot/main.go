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
	"math/rand"
	"strings"
	"time"

	"github.com/ontology-community/onRobot/internal/robot/conf"

	_ "github.com/ontology-community/onRobot/internal/robot"
	core "github.com/ontology-community/onRobot/pkg/frame"
)

var (
	Config               string //config file
	LogConfig            string //Log config file
	TestCaseConfig       string // Test case file dir
	WalletConfig         string // Wallet path
	TransferWalletConfig string // Transfer wallet path
	Methods              string //Methods list in cmdline
)

func init() {
	flag.StringVar(&Config, "config", "target/robot/config.json", "Config of ontology-tool")
	flag.StringVar(&TestCaseConfig, "params", "target/robot/params", "Test params")
	flag.StringVar(&WalletConfig, "wallet", "target/robot/wallet.dat", "Wallet path")
	flag.StringVar(&TransferWalletConfig, "transfer", "target/robot/transfer_wallet.dat", "Transfer wallet path")
	flag.StringVar(&Methods, "t", "subnetDelMember", "methods to run. use ',' to split methods")
	flag.Parse()
}

func main() {
	rand.Seed(time.Now().UnixNano())
	conf.SetParamsDir(TestCaseConfig)
	conf.SetWalletPath(WalletConfig)
	conf.SetTransferWalletPath(TransferWalletConfig)
	defer time.Sleep(time.Second)

	err := conf.DefConfig.Init(Config)
	if err != nil {
		panic(err)
	}

	methods := make([]string, 0)
	if Methods != "" {
		methods = strings.Split(Methods, ",")
	}

	core.OntTool.Start(methods)
}
