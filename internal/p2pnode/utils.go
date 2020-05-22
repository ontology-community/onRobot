package p2pnode

import (
	log4 "github.com/alecthomas/log4go"
	"github.com/ontology-community/onRobot/internal/robot/conf"
	"github.com/ontology-community/onRobot/pkg/p2pserver/net/netserver"
	"github.com/ontology-community/onRobot/pkg/p2pserver/net/protocol"
)

func GenerateNetServerWithContinuePort(protocol p2p.Protocol, port uint16) (ns *netserver.NetServer) {
	var err error

	conf.DefConfig.Net.NodePort = port
	if ns, err = netserver.NewNetServer(protocol, conf.DefConfig.Net); err != nil {
		log4.Crashf("[NewNetServer] crashed, err %s", err)
		return nil
	}
	if err = ns.Start(); err != nil {
		log4.Crashf("start netserver failed, err %s", err)
	}
	return
}
