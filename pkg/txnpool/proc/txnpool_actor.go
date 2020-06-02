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

package proc

import (
	"github.com/ontology-community/onRobot/pkg/p2pserver/message/msg_pack"
	"reflect"

	"github.com/ontio/ontology-eventbus/actor"

	log4 "github.com/alecthomas/log4go"
	"github.com/ontio/ontology/core/types"
	tc "github.com/ontio/ontology/txnpool/common"
)

// NewTxActor creates an actor to handle the transaction-based messages from
// network and http
func NewTxActor(s *TXPoolServer) *TxActor {
	a := &TxActor{}
	a.setServer(s)
	return a
}

// TxnActor: Handle the low priority msg from P2P and API
type TxActor struct {
	server *TXPoolServer
}

// Receive implements the actor interface
func (ta *TxActor) Receive(context actor.Context) {
	switch msg := context.Message().(type) {
	case *actor.Started:
		log4.Info("txpool-tx actor started and be ready to receive tx msg")

	case *actor.Stopping:
		log4.Warn("txpool-tx actor stopping")

	case *actor.Restarting:
		log4.Warn("txpool-tx actor restarting")

	case *tc.TxReq:
		sender := msg.Sender

		log4.Debug("txpool-tx actor receives tx from %v ", sender.Sender())

		ta.handleTransaction(msg.Tx)

	default:
		log4.Debug("txpool-tx actor: unknown msg %v type %v", msg, reflect.TypeOf(msg))
	}
}

func (ta *TxActor) handleTransaction(txn *types.Transaction) {
	if err := ta.server.setTransaction(txn.Hash(), txn); err != nil {
		log4.Error("%s", err)
		return
	}
	hash := txn.Hash()
	log4.Trace("txpool-tx actor received tx %s", hash.ToHexString())
	msg := msgpack.NewTxn(txn)
	go ta.server.Net.Broadcast(msg)
}

func (ta *TxActor) setServer(s *TXPoolServer) {
	ta.server = s
}
