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
	"sync"

	"github.com/ontio/ontology-crypto/keypair"
	vconfig "github.com/ontio/ontology/consensus/vbft/config"
)

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

	return resolver
}

func (self *GovNodeMockResolver) IsGovNode(key keypair.PublicKey) bool {
	self.mu.RLock()
	defer self.mu.RUnlock()

	pubKey := vconfig.PubkeyID(key)
	_, ok := self.govNode[pubKey]
	return ok
}

func (self *GovNodeMockResolver) AddGovNode(key keypair.PublicKey) {
	self.mu.Lock()
	defer self.mu.Unlock()

	pubKey := vconfig.PubkeyID(key)
	self.govNode[pubKey] = struct{}{}
}

func (self *GovNodeMockResolver) DelGovNode(key keypair.PublicKey) {
	self.mu.Lock()
	defer self.mu.Unlock()

	pubKey := vconfig.PubkeyID(key)

	delete(self.govNode, pubKey)
}
