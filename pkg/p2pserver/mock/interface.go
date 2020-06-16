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

package mock

import (
	"net"

	"github.com/ontology-community/onRobot/pkg/p2pserver/common"
	"github.com/ontology-community/onRobot/pkg/p2pserver/connect_controller"
)

type Network interface {
	// NewListener will gen random ip to listen
	NewListener(id common.PeerId) (string, net.Listener)
	// addr: ip:port
	NewListenerWithAddr(id common.PeerId, addr string) net.Listener

	// NewDialer will gen random source IP
	NewDialer(id common.PeerId) connect_controller.Dialer
	NewDialerWithHost(id common.PeerId, host string) connect_controller.Dialer
	AllowConnect(id1, id2 common.PeerId)
	DeliverRate(percent uint)
}
