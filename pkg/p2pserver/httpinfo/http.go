package httpinfo

import (
	"fmt"
	log4 "github.com/alecthomas/log4go"
	"github.com/ontology-community/onRobot/pkg/p2pserver/net/netserver"
	"net/http"
	_ "net/http/pprof"
	"strings"
)

const (
	StatList     = "/stat/list"
	StatClear    = "/stat/clear"
	StatHashList = "/stat/hashlist"
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
	http.HandleFunc(StatHashList, s.handleHashList)
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

func (s *TxInfoServer) handleHashList(w http.ResponseWriter, r *http.Request) {
	st, err := s.svr.GetStat()
	if err != nil {
		errors(w, err)
		return
	}
	hashListStr := r.URL.Query().Get("list")
	list := strings.Split(hashListStr, ",")
	data, err := st.GetAndClearMulti(list)
	if err != nil {
		errors(w, err)
	}
	result(w, data)
}
