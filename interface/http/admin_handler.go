package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"sync/atomic"
	"time"

	resty "github.com/go-resty/resty/v2"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/liuliqiang/xds-demo/internal/storage"
	"github.com/liuliqiang/log4go"
)

type AdminServer struct {
	staticDir string
	is        storage.InstanceStorage
	conn      *websocket.Conn
	initDone  atomic.Value
}

type adminServerOpts struct {
	staticDir  string
	storageDir string
}

func NewAdminServerOpts() *adminServerOpts {
	return &adminServerOpts{}
}

func (o *adminServerOpts) WithStaticDir(p string) *adminServerOpts {
	o.staticDir = p
	return o
}

func (o *adminServerOpts) WithStorageDir(p string) *adminServerOpts {
	o.storageDir = p
	return o
}

func NewAdminServer(opts *adminServerOpts) (as *AdminServer, err error) {
	as = &AdminServer{
		staticDir: opts.staticDir,
	}
	as.initDone.Store(false)

	if as.is, err = storage.NewFileInstanceStorage(opts.storageDir); err != nil {
		return nil, fmt.Errorf("new file storage: %v", err)

	}

	return as, nil
}

func (a *AdminServer) Index(resp http.ResponseWriter, req *http.Request) {
	indexPath := path.Join(a.staticDir, "index.html")
	file, err := os.Open(indexPath)
	if err != nil {
		resp.WriteHeader(http.StatusServiceUnavailable)
		_, _ = resp.Write([]byte(err.Error()))
		return
	}

	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		resp.WriteHeader(http.StatusServiceUnavailable)
		_, _ = resp.Write([]byte(err.Error()))
		return
	}

	_, _ = resp.Write(bytes)
	resp.Header().Set("Content-CMD", "text/html")
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func (s *AdminServer) Websocket(w http.ResponseWriter, r *http.Request) {
	log4go.Info("header origin: " + r.Header.Get("Origin"))
	log4go.Info("request host: " + r.Host)

	if r.Header.Get("Origin") != "http://"+r.Host {
		http.Error(w, "Origin not allowed", 403)
		return
	}
	conn, err := upgrader.Upgrade(w, r, w.Header())
	if err != nil {
		http.Error(w, "Could not open websocket connection", http.StatusBadRequest)
	}

	go s.process(conn)
	go s.checkState()
}

const (
	CMDUpdateHealth = "UPDATE_HEALTH"
	CMDResetCount   = "RESET_COUNT"
	CMDQueryState   = "QUERY_STATES"
	CMDAddInstance  = "ADD_INSTANCE"
)

type CMDRequest struct {
	CMD    string   `json:"cmd"`
	Num    int      `json:"num"`
	Health bool     `json:"health"`
	Host   string   `json:"host"`
	Hosts  []string `json:"hosts"`
}

type CMDResponse struct {
	CMD  string      `json:"cmd"`
	Data interface{} `json:"data"`
}

func (s *AdminServer) process(conn *websocket.Conn) {
	s.conn = conn
	s.initDone.Store(true)
	for {
		msgType, msg, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log4go.Error("Error reading message: %v", err)
			}
			log4go.Info("Failed to read message: %v", err)
			break
		}
		log4go.Debug("Got message type: %#v\n", msgType)
		log4go.Debug("Got message: %#v\n", msg)

		m := CMDRequest{}
		if err = json.Unmarshal(msg, &m); err != nil {
			log4go.Error("Failed to unmarshal msg `%s`: %v", msg, err)
			continue
		}

		result := true
		switch m.CMD {
		case CMDUpdateHealth:
			result = s.updateHealth(&m)
		case CMDResetCount:
			result = s.resetCount(&m)
		case CMDQueryState:
			result = s.queryStates(&m)
		case CMDAddInstance:
			result = s.addInstance(&m)
		default:
			if err = conn.WriteJSON(map[string]interface{}{
				"error": "unsupported type: " + m.CMD,
			}); err != nil {
				log4go.Error("Failed to write out json: %v", err)
			}
		}

		if !result {
			log4go.Info("exit channel")
			return
		}
	}
}

func (s *AdminServer) checkState() {
	tick := time.Tick(time.Second)
	for {
		<-tick

		insts, err := s.is.ListInstance()
		if err != nil {
			log4go.Error("Failed to list instances: %v", err)
			continue
		}

		for _, inst := range insts {
			s.checkInstanceStates(inst)
		}
	}
}

func (s *AdminServer) checkInstanceStates(req *storage.Instance) {
	var needSync = false
	inst, err := s.is.GetInstance(req.Name)
	if err != nil {
		log4go.Warn("Failed to find instant %s: %v", req.Name, err)
		return
	}

	url := fmt.Sprintf("http://%s/ping/state", inst.Host)
	resp, err := resty.New().R().EnableTrace().Get(url)
	stat := instanceStateResp{}
	if err != nil {
		log4go.Error("Failed to get inst state: %v", err)
		if inst.Health {
			needSync = true
		}
		stat = instanceStateResp{
			Name:   req.Name,
			Health: false,
			Count:  req.Count,
		}
	} else {
		if err = json.Unmarshal(resp.Body(), &stat); err != nil {
			log4go.Error("Failed to unmarshal response: %v", err)
			return
		}
	}

	if stat.Name != inst.Name ||
		stat.Health != inst.Health ||
		stat.Count != inst.Count {
		needSync = true
	}

	inst.Name = stat.Name
	inst.Count = stat.Count
	inst.Health = stat.Health
	if err = s.is.AddInstance(inst); err != nil {
		log4go.Error("Failed to update instance: %v", err)
		return
	}

	if needSync && s.initDone.Load().(bool) {
		s.queryStates(&CMDRequest{
			CMD: CMDQueryState,
		})
	}
}

func (s *AdminServer) updateHealth(msg *CMDRequest) bool {
	for _, host := range msg.Hosts {
		inst, err := s.is.GetInstance(host)
		if err != nil {
			log4go.Warn("Failed to find instant: %s", host)
			continue
		}

		url := fmt.Sprintf("http://%s/ping/change_health", inst.Host)
		_, err = resty.New().R().EnableTrace().Get(url)
		if err != nil {
			log4go.Error("Failed to update node helath")
		}
	}

	if err := s.conn.WriteJSON([]byte("{}")); err != nil {
		log4go.Error("Failed to write out json: %v", err)
		return false
	}
	return true
}

func (s *AdminServer) resetCount(msg *CMDRequest) bool {
	for _, host := range msg.Hosts {
		inst, err := s.is.GetInstance(host)
		if err != nil {
			log4go.Warn("Failed to find instant: %s", host)
			continue
		}

		url := fmt.Sprintf("http://%s/ping/reset_count", inst.Host)
		_, err = resty.New().R().EnableTrace().Get(url)
		if err != nil {
			log4go.Error("Failed to reset node %s count", host)
		}
	}

	if err := s.conn.WriteJSON([]byte("{}")); err != nil {
		log4go.Error("Failed to write out json: %v", err)
		return false
	}

	return true
}

func (s *AdminServer) queryStates(req *CMDRequest) bool {
	insts, err := s.is.ListInstance()
	if err != nil {
		log4go.Warn("Failed to list instants: %v", err)
		return true
	}

	resp := CMDResponse{
		CMD:  req.CMD,
		Data: insts,
	}
	if insts == nil {
		resp.Data = []*storage.Instance{}
	}

	if err = s.conn.WriteJSON(resp); err != nil {
		log4go.Error("Failed to write out: %v", err)
		return false
	}

	return true
}

type instanceStateResp struct {
	Name   string `json:"node"`
	Health bool   `json:"health"`
	Count  int    `json:"count"`
}

func (s *AdminServer) addInstance(req *CMDRequest) bool {
	url := fmt.Sprintf("http://%s/ping/state", req.Host)
	resp, err := resty.New().R().EnableTrace().Get(url)
	if err != nil {
		log4go.Error("Failed to get node %s health: %v", req.Host, err)
		return true
	}

	r := instanceStateResp{}
	if err = json.Unmarshal(resp.Body(), &r); err != nil {
		log4go.Warn("Failed to unmarshal instance status resp: %v", err)
		r.Name = uuid.New().String()
	}

	inst := &storage.Instance{
		Name:   r.Name,
		Health: r.Health,
		Host:   req.Host,
		Count:  r.Count,
	}
	if err = s.is.AddInstance(inst); err != nil {
		log4go.Error("Failed to add instance %s: %v", inst.Name, err)
		return true
	}

	cmdResp := CMDResponse{
		CMD:  req.CMD,
		Data: inst,
	}
	if err = s.conn.WriteJSON(cmdResp); err != nil {
		log4go.Error("Failed to write out: %v", err)
		return false
	}

	return true
}
