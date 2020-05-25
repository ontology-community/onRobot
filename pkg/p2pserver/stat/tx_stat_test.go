package stat

import (
	common2 "github.com/ontio/ontology/common"
	types2 "github.com/ontio/ontology/core/types"
	"github.com/ontology-community/onRobot/pkg/p2pserver/common"
	"github.com/ontology-community/onRobot/pkg/p2pserver/message/types"
	"testing"
)

// todo(fuk): delete aftet test
func TestTxStat_HandleRecvMsg(t *testing.T) {
	st := NewMsgStat()

	kid := common.RandPeerKeyId()

	addr, err := common2.AddressFromBase58("AG4pZwKa9cr8ca7PED7FqzUfcwnrQ2N26w")
	if err != nil {
		t.Fatal(err)
	}
	tx := &types.Trn{Txn: &types2.Transaction{
		Version:    byte(1),
		TxType:     types2.InvokeNeo,
		Nonce:      12,
		GasPrice:   0,
		GasLimit:   20000,
		Payer:      addr,
		SignedAddr: []common2.Address{addr},
	}}
	payload := &types.MsgPayload{
		Id:      kid.Id,
		Payload: tx,
	}
	st.HandleRecvMsg(payload)

	hash := tx.Txn.Hash()
	t.Logf("hash %s", hash.ToHexString())

	count := st.RecvMsgCount(hash.ToHexString())
	t.Logf("recv msg %d", count)

	output := st.DumpRecvPeerMsgCountList(hash.ToHexString())

	t.Logf("msg list %s", output)
}

func TestTxStat_HandleSendMsg(t *testing.T) {
	st := NewMsgStat()

	kid := common.RandPeerKeyId()

	addr, err := common2.AddressFromBase58("AG4pZwKa9cr8ca7PED7FqzUfcwnrQ2N26w")
	if err != nil {
		t.Fatal(err)
	}
	tx := &types.Trn{Txn: &types2.Transaction{
		Version:    byte(1),
		TxType:     types2.InvokeNeo,
		Nonce:      12,
		GasPrice:   0,
		GasLimit:   20000,
		Payer:      addr,
		SignedAddr: []common2.Address{addr},
	}}
	st.HandleSendMsg(kid.Id, tx)

	hash := tx.Txn.Hash()
	t.Logf("hash %s", hash.ToHexString())

	count := st.SendMsgCount(hash.ToHexString())
	t.Logf("send msg %d", count)

	output := st.DumpSendPeerMsgCountList(hash.ToHexString())

	t.Logf("msg list %s", output)
}
