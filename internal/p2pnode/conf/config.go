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
	Version   string
	DefConfig = NewDHTConfig()
	//ParamsFileDir string
)

type DHTConfig struct {
	GasPrice uint64
	GasLimit uint64
	Net      *cmf.P2PNodeConfig
	SeedList []string
}

func NewDHTConfig() *DHTConfig {
	return &DHTConfig{}
}

func (c *DHTConfig) Init(fileName string, nodePort, httpInfoPort uint) error {
	err := files.LoadConfig(fileName, c)
	if err != nil {
		return fmt.Errorf("loadConfig error:%s", err)
	}

	c.Net.NodePort = uint16(nodePort)
	c.Net.HttpInfoPort = uint16(httpInfoPort)

	cmf.DefConfig.P2PNode = c.Net
	cmf.DefConfig.Genesis.SeedList = c.SeedList
	cmf.DefConfig.Common.GasPrice = c.GasPrice
	cmf.DefConfig.Common.GasLimit = c.GasLimit

	//ParamsFileDir = paramsDir
	return nil
}
