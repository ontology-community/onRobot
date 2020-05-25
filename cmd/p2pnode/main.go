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
	p2pPort uint
)

func init() {
	flag.StringVar(&Config, "config", "config.json", "Config of ontology-tool")
	flag.StringVar(&LogConfig, "log", "log4go.xml", "Log config of ontology-tool")
	flag.UintVar(&httpPort, "httpport", 10338, "http info port")
	flag.UintVar(&p2pPort, "p2pport", 20338, "http info port")
	flag.Parse()
}

func main() {
	log4.LoadConfiguration(LogConfig)
	if err := conf.DefConfig.Init(Config, p2pPort, httpPort); err != nil {
		log4.Crash(err)
	}
	p2pnode.NewNode()
}
