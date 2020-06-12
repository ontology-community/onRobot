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

package utils

import (
	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/consensus/vbft/config"
	"github.com/ontology-community/onRobot/pkg/p2pserver/storage"
	"sync"
	"time"
)

const GovNodeRefreshDuration = 60

type GovNodeMockResolver struct {
	mu      *sync.RWMutex
	govNode map[string]struct{}
}

// NewGovNodeMockResolver(gov []string)
func NewGovNodeMockResolver() *GovNodeMockResolver {
	govNode := make(map[string]struct{})
	resolver := &GovNodeMockResolver{
		govNode: govNode,
		mu:      new(sync.RWMutex),
	}

	resolver.refresh()

	return resolver
}

func (self *GovNodeMockResolver) refresh() {
	tr := time.NewTimer(0)
	for {
		select {
		case <-tr.C:
			self.cached()
			// todo validate
			tr.Reset(GovNodeRefreshDuration * time.Second)
		default:
		}
	}
}

func (self *GovNodeMockResolver) cached() {
	peers := storage.GetAllSubnet()

	self.mu.Lock()
	for _, node := range peers {
		self.govNode[node] = struct{}{}
	}
	self.mu.Unlock()
}

func (self *GovNodeMockResolver) IsGovNode(key keypair.PublicKey) bool {
	pubKey := vconfig.PubkeyID(key)

	self.mu.RLock()
	defer self.mu.RUnlock()

	_, ok := self.govNode[pubKey]
	return ok
}
