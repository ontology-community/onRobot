package methods

import (
	"github.com/ontology-community/onRobot/p2pserver/common"
	"testing"
)

// go test -count=1 -v github.com/ontology-community/onRobot/methods -run TestCPL
func TestCPL(t *testing.T) {
	remoteKid := common.RandPeerKeyId()

	// random local kid
	localKid := common.RandPeerKeyId()
	distance := common.CommonPrefixLen(remoteKid.Id, localKid.Id)
	t.Log("random local kid distance", remoteKid.Id.ToUint64(), localKid.Id.ToUint64(), distance)
}

// go test -count=1 -v github.com/ontology-community/onRobot/methods -run TestFakePeerID
func TestFakePeerID(t *testing.T) {
	kid := common.RandPeerKeyId()

	fakePeerIDs, err := GenerateFakePeerIDs(kid.Id, 16)
	if err != nil {
		t.Fatal(err)
	}
	for _, id := range fakePeerIDs {
		t.Log("fake id", id.ToUint64(), distance(kid.Id, id))
	}
}
