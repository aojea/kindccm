package main

import (
	"net"
	"os/exec"
)

type Netconfig struct{}

func (n Netconfig) SetupNetwork(ipNet *net.IPNet, dev string) error {
	if err := exec.Command("ifconfig", dev, "inet", ipNet.IP.String(), ipNet.IP.String(), "up").Run(); err != nil {
		return err
	}
	return nil
}

func (n Netconfig) CreateRoutes(ipNet *net.IPNet, dev string) error {
	if err := exec.Command("route", "-n", "add", ipNet.String(), "-interface", dev).Run(); err != nil {
		return err
	}
	return nil
}

func (n Netconfig) DeleteRoutes(ipNet *net.IPNet, dev string) error {
	if err := exec.Command("route", "-n", "delete", ipNet.String(), "-interface", dev).Run(); err != nil {
		return err
	}
	return nil
}

func NewNetconfig() Netconfig {
	return Netconfig{}
}
