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

package robot

import (
	"math/rand"
	"time"

	core "github.com/ontology-community/onRobot/pkg/frame"
	"github.com/ontology-community/onRobot/pkg/p2pserver/params"
)

func init() {
	rand.Seed(time.Now().UnixNano())
	params.InitializeTestParams()

	core.OntTool.RegGCFunc(reset)

	core.OntTool.RegMethod("demo", Demo)
	core.OntTool.RegMethod("fakePeerID", FakePeerID)
	core.OntTool.RegMethod("connect", Connect)
	core.OntTool.RegMethod("handshakeTimeout", HandshakeTimeout)
	core.OntTool.RegMethod("handshakeWrongMsg", HandshakeWrongMsg)
	core.OntTool.RegMethod("heartbeat", Heartbeat)
	core.OntTool.RegMethod("heartbeatInterruptPing", HeartbeatInterruptPing)
	core.OntTool.RegMethod("heartbeatInterruptPong", HeartbeatInterruptPong)
	core.OntTool.RegMethod("resetPeerID", ResetPeerID)
	core.OntTool.RegMethod("ddos", DDos)
	core.OntTool.RegMethod("askFakeBlocks", AskFakeBlocks)
	core.OntTool.RegMethod("attackTxPool", AttackTxPool)
	core.OntTool.RegMethod("doubleSpend", DoubleSpend)
	core.OntTool.RegMethod("transferOnt", TransferOnt)
	core.OntTool.RegMethod("txCount", TxCount)
	core.OntTool.RegMethod("neighbor", Neighbor)
}
