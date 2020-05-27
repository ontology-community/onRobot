package httpinfo

import (
	"fmt"
	log4 "github.com/alecthomas/log4go"
	"github.com/ontology-community/onRobot/pkg/p2pserver/net/netserver"
	"net/http"
)

const (
	StatList  = "/stat/list"
	StatClear = "/stat/clear"
)

type TxInfoServer struct {
	svr *netserver.NetServer
}

func RunTxInfoHttpServer(srv *netserver.NetServer, port uint16) {
	ts := &TxInfoServer{
		svr: srv,
	}
	ts.HandleHttpServer(port)
}

func (s *TxInfoServer) HandleHttpServer(port uint16) {
	http.HandleFunc(StatList, s.handleStat)
	http.HandleFunc(StatClear, s.handleClear)
	addr := fmt.Sprintf(":%d", port)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log4.Crash(err)
	}
	log4.Info("tx stat info server started, listen on %d!", port)
}

func (s *TxInfoServer) handleStat(w http.ResponseWriter, r *http.Request) {
	st, err := s.svr.GetStat()
	if err != nil {
		errors(w, err)
		return
	}
	data, err := st.Stat()
	if err != nil {
		errors(w, err)
		return
	}
	result(w, data)
}

func (s *TxInfoServer) handleClear(w http.ResponseWriter, r *http.Request) {
	st, err := s.svr.GetStat()
	if err != nil {
		errors(w, err)
		return
	}
	st.Clear()
	result(w, true)
}
