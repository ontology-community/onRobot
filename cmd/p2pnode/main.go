package main

import (
	"flag"
	log4 "github.com/alecthomas/log4go"
	"github.com/ontology-community/onRobot/internal/p2pnode"
	"github.com/ontology-community/onRobot/internal/p2pnode/conf"
)

var (
	Config,
	LogConfig string

	httpPort,
	nodePort uint
)

func init() {
	flag.StringVar(&Config, "config", "target/node/config.json", "Config of ontology-tool")
	flag.StringVar(&LogConfig, "log", "target/node/log4go.xml", "Log config of ontology-tool")
	flag.UintVar(&httpPort, "httpport", 10031, "http info port")
	flag.UintVar(&nodePort, "nodeport", 20031, "http info port")
	flag.Parse()
}

func main() {
	log4.LoadConfiguration(LogConfig)
	if err := conf.DefConfig.Init(Config, nodePort, httpPort); err != nil {
		log4.Crash(err)
	}
	p2pnode.NewNode()
}
