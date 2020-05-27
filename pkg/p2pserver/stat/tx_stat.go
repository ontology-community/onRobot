package stat

import (
	"github.com/ontology-community/onRobot/pkg/p2pserver/common"
	"github.com/ontology-community/onRobot/pkg/p2pserver/message/types"
	"sync"
)

type TxStat struct {
	sendmap,
	recvmap map[string]uint64

	smu,
	rmu *sync.RWMutex
}

func NewMsgStat() *TxStat {
	st := &TxStat{}
	st.smu = new(sync.RWMutex)
	st.rmu = new(sync.RWMutex)
	st.sendmap = make(map[string]uint64)
	st.recvmap = make(map[string]uint64)
	return st
}

func (s *TxStat) HandleSendMsg(peerId common.PeerId, message types.Message) {
	if message.CmdType() == common.TX_TYPE {
		s.smu.Lock()
		if tx, ok := message.(*types.Trn); ok {
			hash := tx.Txn.Hash()
			s.sendmap[hash.ToHexString()] += 1
		}
		s.smu.Unlock()
	}
}

func (s *TxStat) HandleRecvMsg(payload *types.MsgPayload) {
	if payload.Payload.CmdType() == common.TX_TYPE {
		s.rmu.Lock()
		message := payload.Payload
		if tx, ok := message.(*types.Trn); ok {
			hash := tx.Txn.Hash()
			s.recvmap[hash.ToHexString()] += 1
		}
		s.rmu.Unlock()
	}
}

func (s *TxStat) SendMsgCount() map[string]uint64 {
	s.smu.RLock()
	defer s.smu.RUnlock()
	return s.sendmap
}

func (s *TxStat) RecvMsgCount() map[string]uint64 {
	s.rmu.RLock()
	defer s.rmu.RUnlock()
	return s.recvmap
}

func (s *TxStat) ClearMsgCount() {
	s.smu.Lock()
	s.sendmap = make(map[string]uint64)
	s.smu.Unlock()

	s.rmu.Lock()
	s.recvmap = make(map[string]uint64)
	s.rmu.Unlock()
}
