package main

import (
	"fmt"
	"os/exec"
	"strings"
)

type Netconfig struct {
	ip    string
	route string
	dev   string
}

func (n Netconfig) SetupNetwork() error {
	sargs := fmt.Sprintf("interface ip set address name=REPLACE_ME source=static addr=REPLACE_ME mask=REPLACE_ME gateway=none")
	args := strings.Split(sargs, " ")
	args[4] = fmt.Sprintf("name=%s", n.dev)
	args[6] = fmt.Sprintf("addr=%s", n.ip)
	// Set a /32 mask because the important is the route through the interface
	args[7] = fmt.Sprintf("mask=255.255.255.255")
	cmd := exec.Command("netsh", args...)
	return cmd.Run()
}

func (n Netconfig) CreateRoutes() error {
	// TODO
	return nil
}

func (n Netconfig) DeleteRoutes() error {
	// TODO
	return nil
}

func NewNetconfig(ip, route, dev string) Netconfig {
	return Netconfig{
		ip:    ip,
		route: route,
		dev:   dev,
	}
}
