package p2pnode

import (
	"fmt"
	log4 "github.com/alecthomas/log4go"
	"github.com/ontio/ontology/events"
	"github.com/ontology-community/onRobot/internal/p2pnode/conf"
	"github.com/ontology-community/onRobot/pkg/p2pserver"
	netreqactor "github.com/ontology-community/onRobot/pkg/p2pserver/actor/req"
	"github.com/ontology-community/onRobot/pkg/p2pserver/httpinfo"
	"github.com/ontology-community/onRobot/pkg/p2pserver/net/netserver"
	"github.com/ontology-community/onRobot/pkg/p2pserver/net/protocol"
	"github.com/ontology-community/onRobot/pkg/p2pserver/protocols"
	"github.com/ontology-community/onRobot/pkg/txnpool"
	"github.com/ontology-community/onRobot/pkg/txnpool/proc"
)

func NewNode() {
	events.Init()
	tp, err := initTxPool()
	if err != nil {
		log4.Crash(err)
	}
	msghandler := initProtocol()

	p2p, err := initP2PNode(tp, msghandler)
	if err != nil {
		panic(err)
	}
	ns := p2p.GetNetwork().(*netserver.NetServer)
	httpinfo.RunTxInfoHttpServer(ns, conf.DefConfig.Net.HttpInfoPort)
}

func initTxPool() (*proc.TXPoolServer, error) {

	txPoolServer, err := txnpool.StartTxnPoolServer()
	if err != nil {
		return nil, fmt.Errorf("Init txpool error: %s", err)
	}

	//hserver.SetTxPid(txPoolServer.GetPID())

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
	netreqactor.SetTxnPoolPid(txpoolSvr.GetPID())
	txpoolSvr.Net = p2p.GetNetwork()
	//hserver.SetNetServer(p2p)
	//p2p.WaitForPeersStart()
	log4.Info("P2P init success")
	return p2p, nil
}
