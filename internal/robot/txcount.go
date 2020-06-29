package robot

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/ontology-community/onRobot/pkg/p2pserver/httpinfo"
	"github.com/ontology-community/onRobot/pkg/p2pserver/stat"

	"github.com/ontio/ontology/account"
	"github.com/ontio/ontology/common/log"
	"github.com/ontology-community/onRobot/internal/robot/conf"
	"github.com/ontology-community/onRobot/pkg/sdk"

	"github.com/ontology-community/onRobot/pkg/p2pserver/message/types"
	"github.com/ontology-community/onRobot/pkg/p2pserver/peer"
	"github.com/ontology-community/onRobot/pkg/p2pserver/protocols"
)

type httpClient struct {
	ips []string
	startHttpPort,
	endHttpPort uint16

	list map[string]*http.Client
}

type WrapTxNum struct {
	stat.TxNum
	Ip   string
	Port uint16
}

func NewHttpClient(ipList []string, startHttpPort, endHttpPort uint16) *httpClient {
	c := &httpClient{}
	c.ips = ipList
	c.startHttpPort = startHttpPort
	c.endHttpPort = endHttpPort
	c.list = make(map[string]*http.Client)

	for _, ip := range c.ips {
		for p := c.startHttpPort; p <= c.endHttpPort; p++ {
			c.setCli(ip, p)
		}
	}
	return c
}

func (c *httpClient) setCli(ip string, port uint16) {
	addr := assembleIpAndPort(ip, port)
	client := new(http.Client)
	client.Timeout = 10 * time.Second
	c.list[addr] = client
}

func (c *httpClient) getCli(ip string, port uint16) (*http.Client, error) {
	addr := assembleIpAndPort(ip, port)
	if cli, ok := c.list[addr]; !ok {
		return nil, fmt.Errorf("http client for %s:%d not exist", ip, port)
	} else {
		return cli, nil
	}
}

func (c *httpClient) statMsgCount() map[string]*stat.TxNum {
	list := make(map[string]*stat.TxNum)

	for _, ip := range c.ips {
		for p := c.startHttpPort; p <= c.endHttpPort; p++ {
			time.Sleep(100 * time.Millisecond)
			data, err := c.getStatResult(ip, p, httpinfo.StatList)
			if err != nil {
				log.Errorf("[send/count] err:%s", err)
			} else {
				for _, v := range data {
					if _, ok := list[v.Hash]; !ok {
						list[v.Hash] = &stat.TxNum{Hash: v.Hash, Send: 0, Recv: 0}
					}
					list[v.Hash].Send += v.Send
					list[v.Hash].Recv += v.Recv
				}
			}
		}
	}

	return list
}

func (c *httpClient) clearMsgCount() {
	log.Debug("clear msg stat")

	for _, ip := range c.ips {
		for p := c.startHttpPort; p <= c.endHttpPort; p++ {
			time.Sleep(100 * time.Millisecond)
			if err := c.getClearResult(ip, p, httpinfo.StatClear); err != nil {
				log.Errorf("[send/count] err:%s", err)
			}
		}
	}
}

func (c *httpClient) statHashList(hashlist []string) []*WrapTxNum {
	list := make([]*WrapTxNum, 0)

	for _, ip := range c.ips {
		for p := c.startHttpPort; p <= c.endHttpPort; p++ {
			time.Sleep(100 * time.Millisecond)
			data, err := c.getHashListResult(ip, p, httpinfo.StatHashList, hashlist)
			if err != nil {
				log.Errorf("[send/count] err:%s", err)
			} else {
				for _, v := range data {
					tx := &WrapTxNum{Ip: ip, Port: p, TxNum: stat.TxNum{
						Hash: v.Hash, Send: v.Send, Recv: v.Recv,
					}}
					list = append(list, tx)
				}
			}
		}
	}

	return list
}

func (c *httpClient) getStatResult(ip string, port uint16, method string) (count []*stat.TxNum, err error) {
	var (
		req  *http.Request
		resp *http.Response
		res  = &httpinfo.Resp{}
		data string
		cli  *http.Client
		ok   bool
	)

	url := fmt.Sprintf("http://%s:%d%s", ip, port, method)
	if req, err = http.NewRequest("GET", url, nil); err != nil {
		return
	}
	if cli, err = c.getCli(ip, port); err != nil {
		return
	}
	if resp, err = cli.Do(req); err != nil {
		return
	}
	defer resp.Body.Close()
	if err = parseResponse(resp.Body, res); err != nil {
		return
	}
	if res.Succeed == false {
		err = fmt.Errorf("%s", res.Err)
		return
	}
	if data, ok = res.Data.(string); !ok {
		err = fmt.Errorf("stat count type invalid")
	}
	count, err = stat.Parse2TxNumList(data)
	return
}

func (c *httpClient) getClearResult(ip string, port uint16, method string) (err error) {
	var (
		req  *http.Request
		resp *http.Response
		res  = &httpinfo.Resp{}
		cli  *http.Client
	)

	url := fmt.Sprintf("http://%s:%d%s", ip, port, method)
	if req, err = http.NewRequest("GET", url, nil); err != nil {
		return
	}
	if cli, err = c.getCli(ip, port); err != nil {
		return
	}
	if resp, err = cli.Do(req); err != nil {
		return
	}
	defer resp.Body.Close()
	if err = parseResponse(resp.Body, res); err != nil {
		return
	}

	if res.Succeed == false {
		err = fmt.Errorf("%s", res.Err)
		return
	}

	return
}

func (c *httpClient) getHashListResult(ip string, port uint16, method string, hashlist []string) (count []*stat.TxNum, err error) {
	var (
		req  *http.Request
		resp *http.Response
		res  = &httpinfo.Resp{}
		data string
		cli  *http.Client
		ok   bool
	)

	reqparam := strings.Join(hashlist, ",")
	url := fmt.Sprintf("http://%s:%d%s?list=%s", ip, port, method, reqparam)
	if req, err = http.NewRequest("GET", url, nil); err != nil {
		return
	}
	if cli, err = c.getCli(ip, port); err != nil {
		return
	}
	if resp, err = cli.Do(req); err != nil {
		return
	}
	defer resp.Body.Close()
	if err = parseResponse(resp.Body, res); err != nil {
		return
	}
	if res.Succeed == false {
		err = fmt.Errorf("%s", res.Err)
		return
	}
	if data, ok = res.Data.(string); !ok {
		err = fmt.Errorf("stat count type invalid")
	}
	count, err = stat.Parse2TxNumList(data)
	return
}

func assembleIpAndPort(ip string, port uint16) string {
	return fmt.Sprintf("%s:%d", ip, port)
}

func parseResponse(body io.Reader, res interface{}) error {
	data, err := ioutil.ReadAll(body)
	if err != nil {
		return fmt.Errorf("read http body error:%s", err)
	}

	err = json.Unmarshal(data, res)
	if err != nil {
		return fmt.Errorf("json.Unmarshal RestfulResp:%s error:%s", body, err)
	}

	return nil
}

type invalidTxWorker struct {
	pr   *peer.Peer
	acc  *account.Account
	list map[string]struct{}
	mu   *sync.RWMutex
}

func NewInvalidTxWorker(remote string) (*invalidTxWorker, error) {
	protocol := protocols.NewHeartbeatHandler()
	ns := GenerateNetServerWithProtocol(protocol)
	pr, err := ns.ConnectAndReturnPeer(remote)
	if err != nil {
		return nil, err
	}
	acc, err := sdk.RecoverAccount(conf.WalletPath, conf.DefConfig.WalletPwd)
	if err != nil {
		return nil, err
	}
	return &invalidTxWorker{
		pr:   pr,
		acc:  acc,
		list: make(map[string]struct{}),
		mu:   new(sync.RWMutex),
	}, nil
}

func (w *invalidTxWorker) sendMultiInvalidTx(num int, dest string) {
	for i := 0; i < num; i++ {
		go w.sendInvalidTxWithoutCheckBalance(dest)
	}
}

func (w *invalidTxWorker) sendInvalidTxWithoutCheckBalance(dest string) (err error) {
	var (
		tran   *types.Trn
		amount uint64 = math.MaxUint64
	)

	if tran, err = GenerateTransferOntTx(w.acc, dest, amount); err != nil {
		return
	}

	// send dump tx
	if err = w.pr.Send(tran); err != nil {
		return
	}

	// save in list
	hash := tran.Txn.Hash()
	w.mu.Lock()
	w.list[hash.ToHexString()] = struct{}{}
	w.mu.Unlock()

	return
}

func (w *invalidTxWorker) getHashList(length int) []string {
	w.mu.Lock()
	defer w.mu.Unlock()
	list := make([]string, 0)
	for v, _ := range w.list {
		list = append(list, v)
		delete(w.list, v)
		if len(list) >= length {
			break
		}
	}
	return list
}

func (w *invalidTxWorker) getMsgCount() int64 {
	w.mu.RLock()
	defer w.mu.RUnlock()
	num := len(w.list)
	return int64(num)
}
