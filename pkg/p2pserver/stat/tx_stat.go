package stat

import (
	"github.com/alecthomas/log4go"
	"github.com/ontology-community/onRobot/pkg/p2pserver/common"
	"github.com/ontology-community/onRobot/pkg/p2pserver/message/types"
	"sync"
)

type TxStat struct {
	smu, rmu   *sync.RWMutex
	send, recv uint64
}

func NewMsgStat() *TxStat {
	st := &TxStat{}
	st.send = 0
	st.recv = 0
	st.smu = new(sync.RWMutex)
	st.rmu = new(sync.RWMutex)
	return st
}

func (s *TxStat) HandleSendMsg(peerId common.PeerId, message types.Message) {
	if message.CmdType() == common.TX_TYPE {
		s.smu.Lock()
		s.send += 1
		s.smu.Unlock()
		log4go.Debug("send msg count %d", s.send)
	}
}

func (s *TxStat) HandleRecvMsg(payload *types.MsgPayload) {
	if payload.Payload.CmdType() == common.TX_TYPE {
		s.rmu.Lock()
		s.recv += 1
		s.rmu.Unlock()
		log4go.Debug("recv msg count %d", s.send)
	}
}

func (s *TxStat) SendMsgCount() uint64 {
	s.smu.RLock()
	defer s.smu.RUnlock()
	return s.send
}

func (s *TxStat) RecvMsgCount() uint64 {
	s.rmu.RLock()
	defer s.rmu.RUnlock()
	return s.recv
}

func (s *TxStat) ClearMsgCount() {
	s.smu.Lock()
	s.send = 0
	s.smu.Unlock()

	s.rmu.Lock()
	s.recv = 0
	s.rmu.Unlock()
}
