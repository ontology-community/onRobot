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

package conf

import (
	"fmt"

	cmf "github.com/ontio/ontology/common/config"
	"github.com/ontology-community/onRobot/pkg/files"
)

var (
	Version            string
	DefConfig          = NewDHTConfig()
	WalletPath         string
	TransferWalletPath string
	ParamsFileDir      string
)

type DHTConfig struct {
	Seed      []string
	Sync      []string
	Net       *cmf.P2PNodeConfig
	Sdk       *SDKConfig
	WalletPwd string
}

type SDKConfig struct {
	JsonRpcAddress   string
	RestfulAddress   string
	WebSocketAddress string

	//Gas Price of transaction
	GasPrice uint64
	//Gas Limit of invoke transaction
	GasLimit uint64
	//Gas Limit of deploy transaction
	GasDeployLimit uint64
}

func NewDHTConfig() *DHTConfig {
	return &DHTConfig{}
}

func (c *DHTConfig) Init(fileName string) error {
	err := files.LoadConfig(fileName, c)
	if err != nil {
		return fmt.Errorf("loadConfig error:%s", err)
	}
	cmf.DefConfig.P2PNode = c.Net
	return nil
}

func SetParamsDir(path string) {
	ParamsFileDir = path
}

func SetWalletPath(path string) {
	WalletPath = path
}

func SetTransferWalletPath(path string) {
	TransferWalletPath = path
}
