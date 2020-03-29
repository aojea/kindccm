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
	if err := exec.Command("ifconfig", n.dev, "inet", n.ip, n.ip, "up").Run(); err != nil {
		return err
	}
	return nil
}

func (n Netconfig) CreateRoutes(dev string) error {
	if n.route == "" {
		return nil
	}
	return exec.Command("route", "-n", "add", n.route, "-interface", dev).Run()
}

func (n Netconfig) DeleteRoutes(dev string) error {
	if n.route == "" {
		return nil
	}
	return exec.Command("route", "-n", "delete", n.route, "-interface", dev).Run()
}

func (n Netconfig) CreateMasquerade(dev string) error {
	// Only for Linux
	return nil
}

func (n Netconfig) DeleteMasquerade(dev string) error {
	// Only for Linux
	return nil
}

func NewNetconfig(ip, route, dev string) Netconfig {
	return Netconfig{
		ip:    ip,
		route: route,
		dev:   dev,
	}
}
