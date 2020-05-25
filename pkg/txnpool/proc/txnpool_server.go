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

// Package proc provides functions for handle messages from
// consensus/ledger/net/http/validators
package proc

import (
	"fmt"
	"github.com/ontio/ontology-eventbus/actor"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/types"
	tc "github.com/ontio/ontology/txnpool/common"
	"github.com/ontology-community/onRobot/pkg/p2pserver/net/protocol"
	"sync"
)

// TXPoolServer contains all api to external modules
type TXPoolServer struct {
	mu     *sync.RWMutex
	txPool map[common.Uint256]*types.Transaction
	actor  *actor.PID // The actors running in the server
	Net    p2p.P2P
}

// NewTxPoolServer creates a new tx pool server to schedule workers to
// handle and filter inbound transactions from the network, http, and consensus.
func NewTxPoolServer() *TXPoolServer {
	s := &TXPoolServer{}
	s.mu = new(sync.RWMutex)
	s.txPool = make(map[common.Uint256]*types.Transaction)
	return s
}

// GetPID returns an actor pid with the actor type, If the type
// doesn't exist, return nil.
func (s *TXPoolServer) GetPID() *actor.PID {
	return s.actor
}

// RegisterActor registers an actor with the actor type and pid.
func (s *TXPoolServer) RegisterActor(actor tc.ActorType, pid *actor.PID) {
	s.actor = pid
}

// UnRegisterActor cancels the actor with the actor type.
func (s *TXPoolServer) UnRegisterActor(actor tc.ActorType) {
	s.actor = nil
}

// Stop stops server and workers.
func (s *TXPoolServer) Stop() {
}

func (s *TXPoolServer) setTransaction(hash common.Uint256, tx *types.Transaction) error {
	s.mu.RLock()
	_, ok := s.txPool[hash]
	s.mu.RUnlock()
	if ok {
		return fmt.Errorf("tx %s already exist", hash.ToHexString())
	}

	s.mu.Lock()
	s.txPool[hash] = tx
	s.mu.Unlock()

	return nil
}
