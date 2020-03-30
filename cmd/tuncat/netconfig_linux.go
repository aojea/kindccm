package main

import (
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

func (n Netconfig) CreateRoutes(dev string) error {
	if n.route == "" {
		return nil
	}
	return exec.Command("ip", "route", "add", n.route, "via", dev).Run()
}

func (n Netconfig) DeleteRoutes(dev string) error {
	if n.route == "" {
		return nil
	}
	return exec.Command("ip", "route", "del", n.route, "via", dev).Run()
}

func (n Netconfig) CreateMasquerade(dev string) error {
	if n.route == "" {
		return nil
	}
	return exec.Command("iptables", "-t", "nat", "-A", "POSTROUTING", "-i", n.dev, "-o", dev, "-j", "MASQUERADE").Run()
}

func (n Netconfig) DeleteMasquerade(dev string) error {
	if n.route == "" {
		return nil
	}
	return exec.Command("iptables", "-t", "nat", "-D", "POSTROUTING", "-i", n.dev, "-o", dev, "-j", "MASQUERADE").Run()
}

func NewNetconfig(ip, route, dev string) Netconfig {
	return Netconfig{
		ip:    ip,
		route: route,
		dev:   dev,
	}
}
