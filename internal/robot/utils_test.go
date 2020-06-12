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
	"testing"

	"github.com/ontology-community/onRobot/pkg/p2pserver/common"
)

// go test -count=1 -v github.com/ontology-community/onRobot/methods -run GenerateZeroDistancePeerIDs
func TestGenerateZeroDistancePeerIDs(t *testing.T) {
	kid := common.RandPeerKeyId()

	fakePeerIDs, err := GenerateZeroDistancePeerIDs(kid.Id, 16)
	if err != nil {
		t.Fatal(err)
	}
	for _, id := range fakePeerIDs {
		t.Log("fake id", id.ToUint64(), distance(kid.Id, id))
	}
}

// go test -count=1 -v github.com/ontology-community/onRobot/methods -run TestCPL
func TestCPL(t *testing.T) {
	countlist := make([]int, 8)
	othercount := 0
	for i := 0; i < 100; i++ {
		remoteKid := common.RandPeerKeyId()
		randLocalKid := common.RandPeerKeyId()
		distance := common.CommonPrefixLen(remoteKid.Id, randLocalKid.Id)
		if distance < 8 {
			countlist[distance]++
		} else {
			othercount++
		}
	}
	t.Logf("random local kid distance, %d %d %d %d %d %d %d %d %d",
		countlist[0], countlist[1], countlist[2], countlist[3], countlist[4], countlist[5], countlist[6], countlist[7], othercount,
	)
}
