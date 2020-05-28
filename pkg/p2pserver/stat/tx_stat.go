package stat

import (
	"encoding/json"
	log4 "github.com/alecthomas/log4go"
	"github.com/ontology-community/onRobot/pkg/p2pserver/common"
	"github.com/ontology-community/onRobot/pkg/p2pserver/message/types"
	"sync"
)

type TxNum struct {
	Hash string `json:"hash"`
	Send uint64 `json:"send"`
	Recv uint64 `json:"recv"`
}

type TxStat struct {
	data map[string]*TxNum
	mu   *sync.RWMutex
}

func NewMsgStat() *TxStat {
	st := &TxStat{}
	st.mu = new(sync.RWMutex)
	st.data = make(map[string]*TxNum)
	return st
}

func (s *TxStat) HandleSendMsg(peerId common.PeerId, message types.Message) {
	if message.CmdType() == common.TX_TYPE {
		if tx, ok := message.(*types.Trn); ok {
			s.mu.Lock()
			h := tx.Txn.Hash()
			hash := h.ToHexString()
			s.GenerateTxNum(hash)
			s.data[hash].Send += 1
			log4.Trace("send received tx %s, count %d", hash, s.data[hash].Send)
			s.mu.Unlock()
		}
	}
}

func (s *TxStat) HandleRecvMsg(payload *types.MsgPayload) {
	if payload.Payload.CmdType() == common.TX_TYPE {
		message := payload.Payload
		if tx, ok := message.(*types.Trn); ok {
			s.mu.Lock()
			h := tx.Txn.Hash()
			hash := h.ToHexString()
			s.GenerateTxNum(hash)
			s.data[hash].Recv += 1
			s.mu.Unlock()
		}
	}
}

func (s *TxStat) Stat() (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*TxNum, 0)
	for _, v := range s.data {
		list = append(list, v)
	}
	bz, err := json.Marshal(list)
	if err != nil {
		return "", err
	}
	return string(bz), nil
}

func (s *TxStat) GetAndClearMulti(hashList []string) (string, error) {
	s.mu.Lock()
	list := make([]*TxNum, 0)
	for _, hash := range hashList {
		if tx, ok := s.data[hash]; ok {
			list = append(list, tx)
			delete(s.data, hash)
		} else {
			list = append(list, &TxNum{Hash: hash, Send: 0, Recv: 0})
		}
	}
	s.mu.Unlock()

	bz, err := json.Marshal(list)
	if err != nil {
		return "", err
	}
	return string(bz), nil
}

func (s *TxStat) Clear() {
	s.mu.Lock()
	s.data = make(map[string]*TxNum)
	s.mu.Unlock()
}

func (s *TxStat) GenerateTxNum(hash string) {
	if _, ok := s.data[hash]; !ok {
		s.data[hash] = &TxNum{Hash: hash, Send: 0, Recv: 0}
	}
}

func Parse2TxNumList(data string) ([]*TxNum, error) {
	list := []*TxNum{}
	err := json.Unmarshal([]byte(data), &list)
	return list, err
}
