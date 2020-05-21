package robot

import (
	"github.com/ontology-community/onRobot/pkg/p2pserver/common"
	"testing"
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
