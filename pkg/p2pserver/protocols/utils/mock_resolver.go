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
	"github.com/ontio/ontology/consensus/vbft/config"
)

type MockGovNodeResolver struct {
	db *MockLedger
}

func NewGovNodeMockResolver(db *MockLedger) *MockGovNodeResolver {
	return &MockGovNodeResolver{
		db: db,
	}
}

func (self *MockGovNodeResolver) IsGovNode(key string) bool {
	if self.db == nil {
		return false
	}
	return self.db.exist(key)
}

func (self *MockGovNodeResolver) IsGovNodePubKey(key keypair.PublicKey) bool {
	return self.db.exist(getKey(key))
}

type MockLedger struct {
	mu  *sync.RWMutex
	gov map[string]struct{}
}

func NewMockLedger() *MockLedger {
	return &MockLedger{
		mu:  new(sync.RWMutex),
		gov: make(map[string]struct{}),
	}
}

func (self *MockLedger) AddGovNode(key keypair.PublicKey) {
	self.mu.Lock()
	pubKey := getKey(key)
	self.gov[pubKey] = struct{}{}
	self.mu.Unlock()
}

func (self *MockLedger) DelGovNode(key keypair.PublicKey) {
	self.mu.Lock()
	pubKey := getKey(key)
	delete(self.gov, pubKey)
	self.mu.Unlock()
}

func (self *MockLedger) exist(key string) bool {
	self.mu.RLock()
	defer self.mu.RUnlock()

	_, ok := self.gov[key]
	return ok
}

func getKey(key keypair.PublicKey) string {
	return vconfig.PubkeyID(key)
}
