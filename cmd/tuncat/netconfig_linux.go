package main

import (
	"net"
	"os/exec"
)

type Netconfig struct {
	ip    string
	route string
	dev   string
}

func (n Netconfig) SetupNetwork() error {
	if err := exec.Command("ip", "link", "set", n.dev, "up").Run(); err != nil {
		return err
	}
	if err := exec.Command("ip", "addr", "add", n.ip, "dev", n.dev).Run(); err != nil {
		return err
	}
	return nil
}

func (n Netconfig) CreateRoutes() error {
	if n.route == "" {
		return nil
	}
	return exec.Command("ip", "route", "add", n.route, "via", n.dev).Run()
}

func (n Netconfig) DeleteRoutes() error {
	if n.route == "" {
		return nil
	}
	return exec.Command("ip", "route", "del", n.route, "via", n.dev).Run()
}

func NewNetconfig(ip, route, dev string) Netconfig {
	return Netconfig{
		ip:    ip,
		route: route,
		dev:   dev,
	}
}
