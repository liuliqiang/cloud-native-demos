package server

import (
	"encoding/json"
	"net/http"
	"strconv"
)

type PingService struct {
	iden   string
	count  int
	health bool
}

func NewPingService(iden string) *PingService {
	return &PingService{
		iden:   iden,
		count:  0,
		health: true,
	}
}

func (s *PingService) Ping(resp http.ResponseWriter, req *http.Request) {
	if !s.health {
		resp.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	s.count++
	msg := req.URL.Query().Get("say")
	_, _ = resp.Write([]byte(s.iden + ": " + msg))
}

func (s *PingService) Health(resp http.ResponseWriter, req *http.Request) {
	if !s.health {
		resp.WriteHeader(http.StatusServiceUnavailable)
		return
	}
}

func (s *PingService) ChangeHealth(resp http.ResponseWriter, req *http.Request) {
	s.health = !s.health
}

func (s *PingService) Count(resp http.ResponseWriter, req *http.Request) {
	_, _ = resp.Write([]byte(strconv.Itoa(s.count)))
}

func (s *PingService) ResetCount(resp http.ResponseWriter, req *http.Request) {
	s.count = 0
}

type NodeState struct {
	Node   string `json:"node"`
	Health bool   `json:"health"`
	Count  int    `json:"count"`
}

func (s *PingService) State(resp http.ResponseWriter, req *http.Request) {
	ns := &NodeState{
		Node:   s.iden,
		Health: s.health,
		Count:  s.count,
	}

	bytes, err := json.Marshal(ns)
	if err != nil {
		resp.WriteHeader(http.StatusServiceUnavailable)
		_, _ = resp.Write([]byte(err.Error()))
		return
	}

	if !s.health {
		resp.WriteHeader(http.StatusServiceUnavailable)
	}
	_, _ = resp.Write(bytes)
}
