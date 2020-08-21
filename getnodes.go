package main

import (
	"encoding/json"
	"io/ioutil"
)

type Auth struct {
	Login    string
	Password string
	Enable   string
}

type Node struct {
	Hostname string
	Addr     string
	Auth     Auth
}

type Nodes struct {
	Nodes []Node
}

func (n *Nodes) GetNodes(invent_file string) error {
	rawDataIn, err := ioutil.ReadFile(invent_file)
	if err != nil {
		return err
	}
	err = json.Unmarshal(rawDataIn, &n.Nodes)
	if err != nil {
		return err
	}
	return err
}
