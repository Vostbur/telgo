package main

import (
	"io/ioutil"
	"strings"
)

type Commands struct {
	Cmd []string
}

func (c *Commands) GetCommangs(f string) error {
	rawDataIn, err := ioutil.ReadFile(f)
	if err != nil {
		return err
	}
	c.Cmd = strings.Split(string(rawDataIn), "\n")
	return err
}
