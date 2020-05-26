package httpinfo

import (
	"fmt"
	log4 "github.com/alecthomas/log4go"
	"github.com/ontology-community/onRobot/pkg/p2pserver/net/netserver"
	"net/http"
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

	http.HandleFunc("/stat/send", s.handleSendCount)
	http.HandleFunc("/stat/recv", s.handleRecvCount)
	http.HandleFunc("/stat/clear", s.handleClearCount)
	addr := fmt.Sprintf(":%d", port)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log4.Crash(err)
	}
}

func (s *TxInfoServer) handleSendCount(w http.ResponseWriter, r *http.Request) {
	st, err := s.svr.GetStat()
	if err != nil {
		errors(w, err)
		return
	}
	count := st.SendMsgCount()
	result(w, count)
}

func (s *TxInfoServer) handleRecvCount(w http.ResponseWriter, r *http.Request) {
	st, err := s.svr.GetStat()
	if err != nil {
		errors(w, err)
		return
	}
	count := st.RecvMsgCount()
	result(w, count)
}

func (s *TxInfoServer) handleClearCount(w http.ResponseWriter, r *http.Request) {
	st, err := s.svr.GetStat()
	if err != nil {
		errors(w, err)
		return
	}
	st.ClearMsgCount()
	result(w, true)
}
