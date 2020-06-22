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

package params

import (
	"github.com/ontology-community/onRobot/pkg/p2pserver/common"
	"time"
)

// handshake test cases
const (
	HandshakeNormal = iota
	Handshake_StopClientAfterSendVersion
	Handshake_StopClientAfterReceiveVersion
	Handshake_StopClientAfterUpdateKad
	Handshake_StopClientAfterReadKad
	Handshake_StopClientAfterSendAck
	Handshake_StopServerAfterSendVersion
	Handshake_StopServerAfterReceiveVersion
	Handshake_StopServerAfterUpdateKad
	Handshake_StopServerAfterReadKad
	Handshake_StopServerAfterReadAck
)

var (
	HandshakeLevel                   uint8
	HandshakeWrongMsg                bool
	HandshakeClientTimeout           time.Duration
	HandshakeServerTimeout           time.Duration
	HeartbeatBlockHeight             uint64
	HeartbeatInterruptAfterStartTime int64
	HeartbeatInterruptPingLastTime   int64
	HeartbeatInterruptPongLastTime   int64
	EnableTxStat                     bool
)

var (
	DefHandshakeStopLevel               uint8  = HandshakeNormal
	DefHandshakeWrongMsg                       = false
	DefHandshakeClientTimeout                  = time.Duration(0)
	DefHandshakeServerTimeout                  = time.Duration(0)
	DefHeartbeatBlockHeight             uint64 = 9442
	DefHeartbeatInterruptAfterStartTime int64  = 0
	DefHeartbeatInterruptPingLastTime   int64  = 0
	DefHeartbeatInterruptPongLastTime   int64  = 0
	DefEnableTxStat                     bool   = false
	DefDifficulty                       int    = common.Difficulty
)

func InitializeTestParams() {
	HandshakeLevel = DefHandshakeStopLevel
	HandshakeWrongMsg = DefHandshakeWrongMsg
	HandshakeClientTimeout = DefHandshakeClientTimeout
	HandshakeServerTimeout = DefHandshakeServerTimeout
	HeartbeatBlockHeight = DefHeartbeatBlockHeight
	HeartbeatInterruptAfterStartTime = DefHeartbeatInterruptAfterStartTime
	HeartbeatInterruptPingLastTime = DefHeartbeatInterruptPingLastTime
	HeartbeatInterruptPongLastTime = DefHeartbeatInterruptPongLastTime
	EnableTxStat = DefEnableTxStat
	common.Difficulty = DefDifficulty
}

func Reset() {
	InitializeTestParams()
}

// handshake stop level
func SetHandshakeStopLevel(lvl uint8) {
	HandshakeLevel = lvl
}
func ValidateHandshakeStopLevel(lvl uint8) bool {
	return HandshakeLevel == lvl
}

// handshake wrong msg
func SetHandshakeWrongMsg(active bool) {
	HandshakeWrongMsg = active
}

// handshake timeout
func SetHandshakeClientTimeout(sec int) {
	HandshakeClientTimeout = time.Duration(sec) * time.Second
}
func SetHandshakeServerTimeout(sec int) {
	HandshakeServerTimeout = time.Duration(sec) * time.Second
}

// heartbeat
func SetHeartbeatTestBlockHeight(height uint64) {
	HeartbeatBlockHeight = height
}
func SetHeartbeatTestInterruptAfterStartTime(sec int64) {
	HeartbeatInterruptAfterStartTime = sec
}
func SetHeartbeatTestInterruptPingLastTime(sec int64) {
	HeartbeatInterruptPingLastTime = sec
}
func SetHeartbeatTestInterruptPongLastTime(sec int64) {
	HeartbeatInterruptPongLastTime = sec
}

// tx stat
func SetEnableTxStat() {
	EnableTxStat = true
}

func SetDifficulty(difficulty int) {
	common.Difficulty = difficulty
}
