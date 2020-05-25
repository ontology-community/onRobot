package stat

import (
	"github.com/alecthomas/log4go"
	"github.com/ontology-community/onRobot/pkg/p2pserver/common"
	"github.com/ontology-community/onRobot/pkg/p2pserver/message/types"
	"sync/atomic"
)

type TxStat struct {
	//send,
	//recv map[common.PeerId]map[string]uint
	//send, recv map[common2.Uint256]uint
	//
	//sdmu,
	//rvmu *sync.RWMutex

	send, recv uint64
}

func NewMsgStat() *TxStat {
	st := &TxStat{}
	//st.send = make(map[common.PeerId]map[string]uint)
	//st.recv = make(map[common.PeerId]map[string]uint)
	//st.sdmu = new(sync.RWMutex)
	//st.rvmu = new(sync.RWMutex)
	st.send = 0
	st.recv = 0
	return st
}

func (s *TxStat) HandleSendMsg(peerId common.PeerId, message types.Message) {
	if message.CmdType() == common.TX_TYPE {
		atomic.AddUint64(&s.send, 1)
	}
	log4go.Debug("send msg count %d", s.send)
	//if _, ok := message.(*types2.Transaction); ok {

	//s.sdmu.Lock()
	//var stat map[string]uint
	//stat, ok := s.send[peerId]
	//if !ok {
	//	stat = make(map[string]uint)
	//}
	//hash := data.Txn.Hash()
	//stat[hash.ToHexString()] += 1
	//s.send[peerId] = stat
	//s.sdmu.Unlock()
	//}
}

func (s *TxStat) HandleRecvMsg(payload *types.MsgPayload) {
	if payload.Payload.CmdType() == common.TX_TYPE {
		atomic.AddUint64(&s.recv, 1)
	}
	log4go.Debug("recv msg count %d", s.send)
	//s.rvmu.Lock()
	//peerId := payload.Id
	//msg := payload.Payload
	//if data, ok := msg.(*types.Trn); ok {
	//	var stat map[string]uint
	//	stat, ok := s.recv[peerId]
	//	if !ok {
	//		stat = make(map[string]uint)
	//	}
	//	hash := data.Txn.Hash()
	//	stat[hash.ToHexString()] += 1
	//	s.recv[peerId] = stat
	//}
	//s.rvmu.Unlock()
}

func (s *TxStat) SendMsgCount() uint64 {
	//s.sdmu.RLock()
	//defer s.sdmu.RUnlock()
	//
	//var count uint
	//for _, stat := range s.send {
	//	if n, ok := stat[hash]; ok {
	//		count += n
	//	}
	//}
	return s.send
}

//func (s *TxStat) DumpSendPeerMsgCountList(hash string) string {
//	s.sdmu.RLock()
//	defer s.sdmu.RUnlock()
//
//	var output string
//	for peerID, stat := range s.send {
//		if n, ok := stat[hash]; ok {
//			output += fmt.Sprintf("peer:%s|sendmsg:%d,", peerID.ToHexString(), n)
//		}
//	}
//
//	return output
//}

func (s *TxStat) RecvMsgCount() uint64 {
	/*s.rvmu.RLock()
	defer s.rvmu.RUnlock()

	var count uint
	for _, stat := range s.recv {
		if n, ok := stat[hash]; ok {
			count += n
		}
	}
	return count*/
	return s.recv
}

//func (s *TxStat) DumpRecvPeerMsgCountList(hash string) string {
//	s.rvmu.RLock()
//	defer s.rvmu.RUnlock()
//
//	var output string
//	for peerID, stat := range s.recv {
//		if n, ok := stat[hash]; ok {
//			output += fmt.Sprintf("peer:%s|recvmsg:%d,", peerID.ToHexString(), n)
//		}
//	}
//	return output
//}
