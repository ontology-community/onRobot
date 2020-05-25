package httpinfo

import (
	"fmt"
	log4 "github.com/alecthomas/log4go"
	"github.com/ontology-community/onRobot/pkg/p2pserver/net/netserver"
	"net/http"
)

var (
	ErrParamsInvalid = fmt.Errorf("params invalid")
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
	//http.HandleFunc("/stat/send/dump", s.handleSendDump)
	//http.HandleFunc("/stat/recv/dump", s.handleRecvDump)

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
	//hash, err := getParamsHash(r)
	//if err != nil {
	//	errors(w, err)
	//	return
	//}
	count := st.SendMsgCount()
	result(w, count)
}

//func (s *TxInfoServer) handleSendDump(w http.ResponseWriter, r *http.Request) {
//	st, err := s.svr.GetStat()
//	if err != nil {
//		errors(w, err)
//		return
//	}
//	hash, err := getParamsHash(r)
//	if err != nil {
//		errors(w, err)
//		return
//	}
//	list := st.DumpSendPeerMsgCountList(hash)
//	result(w, list)
//}

func (s *TxInfoServer) handleRecvCount(w http.ResponseWriter, r *http.Request) {
	st, err := s.svr.GetStat()
	if err != nil {
		errors(w, err)
		return
	}
	//hash, err := getParamsHash(r)
	//if err != nil {
	//	errors(w, err)
	//	return
	//}
	count := st.RecvMsgCount()
	result(w, count)
}

//func (s *TxInfoServer) handleRecvDump(w http.ResponseWriter, r *http.Request) {
//	st, err := s.svr.GetStat()
//	if err != nil {
//		errors(w, err)
//		return
//	}
//	hash, err := getParamsHash(r)
//	if err != nil {
//		errors(w, err)
//		return
//	}
//	list := st.DumpSendPeerMsgCountList(hash)
//	result(w, list)
//}

func getParamsHash(r *http.Request) (string, error) {
	vs, ok := r.URL.Query()["hash"]
	if !ok || len(vs) == 0 {
		return "", ErrParamsInvalid
	}
	return vs[0], nil
}
