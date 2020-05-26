package stat

import (
	"github.com/alecthomas/log4go"
	"github.com/ontology-community/onRobot/pkg/p2pserver/common"
	"github.com/ontology-community/onRobot/pkg/p2pserver/message/types"
	"sync/atomic"
)

type TxStat struct {
	send, recv uint64
}

func NewMsgStat() *TxStat {
	st := &TxStat{}
	st.send = 0
	st.recv = 0
	return st
}

func (s *TxStat) HandleSendMsg(peerId common.PeerId, message types.Message) {
	if message.CmdType() == common.TX_TYPE {
		atomic.AddUint64(&s.send, 1)
		log4go.Debug("send msg count %d", s.send)
	}
}

func (s *TxStat) HandleRecvMsg(payload *types.MsgPayload) {
	if payload.Payload.CmdType() == common.TX_TYPE {
		atomic.AddUint64(&s.recv, 1)
		log4go.Debug("recv msg count %d", s.send)
	}
}

func (s *TxStat) SendMsgCount() uint64 {
	return s.send
}

func (s *TxStat) RecvMsgCount() uint64 {
	return s.recv
}
