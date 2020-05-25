package p2pnode

import (
	"fmt"
	log4 "github.com/alecthomas/log4go"
	"github.com/ontio/ontology/events"
	bactor "github.com/ontio/ontology/http/base/actor"
	hserver "github.com/ontio/ontology/http/base/actor"
	"github.com/ontio/ontology/validator/stateful"
	"github.com/ontio/ontology/validator/stateless"
	"github.com/ontology-community/onRobot/internal/p2pnode/conf"
	"github.com/ontology-community/onRobot/pkg/p2pserver"
	netreqactor "github.com/ontology-community/onRobot/pkg/p2pserver/actor/req"
	"github.com/ontology-community/onRobot/pkg/p2pserver/httpinfo"
	"github.com/ontology-community/onRobot/pkg/p2pserver/net/netserver"
	"github.com/ontology-community/onRobot/pkg/p2pserver/net/protocol"
	"github.com/ontology-community/onRobot/pkg/p2pserver/protocols"
	"github.com/ontology-community/onRobot/pkg/txnpool"
	tc "github.com/ontology-community/onRobot/pkg/txnpool/common"
	"github.com/ontology-community/onRobot/pkg/txnpool/proc"
)

const (
	disablePreExec        = true
	disableBroadcastNetTx = false
	disableSyncVerifyTx   = true
)

// todo(fukun):
func NewNode() {
	events.Init()
	tp, err := initTxPool()
	if err != nil {
		log4.Crash(err)
	}
	msghandler := initProtocol()

	p2p, err := initP2PNode(tp, msghandler)
	if err != nil {
		log4.Crash(err)
	}
	ns := p2p.GetNetwork().(*netserver.NetServer)
	httpinfo.RunTxInfoHttpServer(ns, conf.DefConfig.Net.HttpInfoPort)
}

func initTxPool() (*proc.TXPoolServer, error) {
	bactor.DisableSyncVerifyTx = disableSyncVerifyTx
	txPoolServer, err := txnpool.StartTxnPoolServer(disablePreExec, disableBroadcastNetTx)
	if err != nil {
		return nil, fmt.Errorf("Init txpool error: %s", err)
	}
	stlValidator, _ := stateless.NewValidator("stateless_validator")
	stlValidator.Register(txPoolServer.GetPID(tc.VerifyRspActor))
	stlValidator2, _ := stateless.NewValidator("stateless_validator2")
	stlValidator2.Register(txPoolServer.GetPID(tc.VerifyRspActor))
	stfValidator, _ := stateful.NewValidator("stateful_validator")
	stfValidator.Register(txPoolServer.GetPID(tc.VerifyRspActor))

	hserver.SetTxnPoolPid(txPoolServer.GetPID(tc.TxPoolActor))
	hserver.SetTxPid(txPoolServer.GetPID(tc.TxActor))

	log4.Info("TxPool init success")
	return txPoolServer, nil
}

func initProtocol() p2p.Protocol {
	return protocols.NewWithoutBlockSyncMsgHandler()
}

func initP2PNode(txpoolSvr *proc.TXPoolServer, handler p2p.Protocol) (*p2pserver.P2PServer, error) {
	p2p, err := p2pserver.NewStatServer(handler, conf.DefConfig.Net)
	if err != nil {
		return nil, err
	}
	if err := p2p.Start(); err != nil {
		return nil, fmt.Errorf("p2p service start error %s", err)
	}

	netreqactor.SetTxnPoolPid(txpoolSvr.GetPID(tc.TxActor))
	txpoolSvr.Net = p2p.GetNetwork()
	// hserver.SetNetServer(p2p.GetNetwork())
	p2p.WaitForPeersStart()
	log4.Info("P2P init success")
	return p2p, nil
}
