package model

type Instance struct {
	Name   string `json:"name"`
	Health bool   `json:"health"`
	Host   string `json:"host"`
	Count  int    `json:"count"`
}
