package robot

import (
	"fmt"
	"math/rand"

	ontcm "github.com/ontio/ontology/common"
	p2pcm "github.com/ontology-community/onRobot/pkg/p2pserver/common"
)

// GenerateZeroDistancePeerIDs 生成距离为0的peerID列表
func GenerateZeroDistancePeerIDs(tgID p2pcm.PeerId, num int) ([]p2pcm.PeerId, error) {
	if num >= 128 {
		return nil, fmt.Errorf("list length should < 128")
	}

	list := make([]p2pcm.PeerId, 0, num)
	exists := make(map[uint64]struct{})

	sink := new(ontcm.ZeroCopySink)
	tgID.Serialization(sink)
	exists[tgID.ToUint64()] = struct{}{}

	var getValidXorByte = func(tg uint8) uint8 {
		var xor uint8
		for {
			delta := uint8(rand.Int63n(255))
			xor = delta ^ tg
			if xor >= 128 && xor <= 255 {
				break
			}
		}
		return xor
	}

	sinkbz := sink.Bytes()
	for {
		bz := new([20]byte)
		copy(bz[:], sinkbz[:])

		xor := getValidXorByte(sinkbz[0])
		bz[0] = xor

		source := ontcm.NewZeroCopySource(bz[:])
		peerID := p2pcm.PeerId{}
		if err := peerID.Deserialization(source); err != nil {
			continue
		}
		if _, exist := exists[peerID.ToUint64()]; exist {
			continue
		} else {
			exists[peerID.ToUint64()] = struct{}{}
			list = append(list, peerID)
			distance(peerID, tgID)
		}
		if len(list) >= num {
			break
		}
	}

	return list, nil
}

func distance(local, target p2pcm.PeerId) int {
	return p2pcm.CommonPrefixLen(local, target)
}
