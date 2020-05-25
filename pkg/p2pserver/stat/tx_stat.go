package stat

import (
	"fmt"
	"github.com/ontology-community/onRobot/pkg/p2pserver/common"
	"github.com/ontology-community/onRobot/pkg/p2pserver/message/types"
	"sync"
)

type TxStat struct {
	send,
	recv map[common.PeerId]map[string]uint

	sdmu,
	rvmu *sync.RWMutex
}

func NewMsgStat() *TxStat {
	st := &TxStat{}
	st.send = make(map[common.PeerId]map[string]uint)
	st.recv = make(map[common.PeerId]map[string]uint)
	st.sdmu = new(sync.RWMutex)
	st.rvmu = new(sync.RWMutex)
	return st
}

func (s *TxStat) HandleRecvMsg(payload *types.MsgPayload) {
	s.rvmu.Lock()
	peerId := payload.Id
	msg := payload.Payload
	if data, ok := msg.(*types.Trn); ok {
		var stat map[string]uint
		stat, ok := s.recv[peerId]
		if !ok {
			stat = make(map[string]uint)
		}
		hash := data.Txn.Hash()
		stat[hash.ToHexString()] += 1
		s.recv[peerId] = stat
	}
	s.rvmu.Unlock()
}

func (s *TxStat) HandleSendMsg(peerId common.PeerId, message types.Message) {
	if data, ok := message.(*types.Trn); ok {
		s.sdmu.Lock()
		var stat map[string]uint
		stat, ok := s.send[peerId]
		if !ok {
			stat = make(map[string]uint)
		}
		hash := data.Txn.Hash()
		stat[hash.ToHexString()] += 1
		s.send[peerId] = stat
		s.sdmu.Unlock()
	}
}

func (s *TxStat) SendMsgCount(hash string) uint {
	s.sdmu.RLock()
	defer s.sdmu.RUnlock()

	var count uint
	for _, stat := range s.send {
		if n, ok := stat[hash]; ok {
			count += n
		}
	}
	return count
}

func (s *TxStat) DumpSendPeerMsgCountList(hash string) string {
	s.sdmu.RLock()
	defer s.sdmu.RUnlock()

	var output string
	for peerID, stat := range s.send {
		if n, ok := stat[hash]; ok {
			output += fmt.Sprintf("peer:%s|sendmsg:%d,", peerID.ToHexString(), n)
		}
	}

	return output
}

func (s *TxStat) RecvMsgCount(hash string) uint {
	s.rvmu.RLock()
	defer s.rvmu.RUnlock()

	var count uint
	for _, stat := range s.recv {
		if n, ok := stat[hash]; ok {
			count += n
		}
	}
	return count
}

func (s *TxStat) DumpRecvPeerMsgCountList(hash string) string {
	s.rvmu.RLock()
	defer s.rvmu.RUnlock()

	var output string
	for peerID, stat := range s.recv {
		if n, ok := stat[hash]; ok {
			output += fmt.Sprintf("peer:%s|recvmsg:%d,", peerID.ToHexString(), n)
		}
	}
	return output
}
