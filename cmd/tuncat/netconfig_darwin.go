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
	if err := exec.Command("route", "-n", "add", ipNet.String(), "-interface", dev).Run(); err != nil {
		return err
	}
	return nil
}

func NewNetconfig() Netconfig {
	return Netconfig{}
}
