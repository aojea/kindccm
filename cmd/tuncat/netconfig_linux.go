package main

import (
	"net"
	"os/exec"
)

type Netconfig struct{}

func (n Netconfig) SetupNetwork(ipNet *net.IPNet, dev string) error {
	if err := exec.Command("ip", "link", "set", dev, "up").Run(); err != nil {
		return err
	}
	if err := exec.Command("ip", "addr", "add", ipNet.String(), "dev", dev).Run(); err != nil {
		return err
	}
	return nil
}

func (n Netconfig) CreateRoutes(ipNet *net.IPNet, dev string) error {
	if err := exec.Command("ip", "route", "add", ipNet.String(), "via", dev).Run(); err != nil {
		return err
	}
	return nil
}

func (n Netconfig) DeleteRoutes(ipNet *net.IPNet, dev string) error {
	if err := exec.Command("ip", "route", "del", ipNet.String(), "via", dev).Run(); err != nil {
		return err
	}
	return nil
}

func NewNetconfig() Netconfig {
	return Netconfig{}
}
