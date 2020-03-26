package main

import (
	"fmt"
	"os/exec"
	"strings"
)

type Netconfig struct{}

func (n Netconfig) SetupNetwork(ipNet *net.IPNet, dev string) error {
	sargs := fmt.Sprintf("interface ip set address name=REPLACE_ME source=static addr=REPLACE_ME mask=REPLACE_ME gateway=none")
	args := strings.Split(sargs, " ")
	args[4] = fmt.Sprintf("name=%s", dev)
	args[6] = fmt.Sprintf("addr=%s", ipNet.IP)
	args[7] = fmt.Sprintf("mask=%d.%d.%d.%d", ipNet.Mask[0], ipNet.Mask[1], ipNet.Mask[2], ipNet.Mask[3])
	cmd := exec.Command("netsh", args...)
	return cmd.Run()
}

func NewNetconfig() Netconfig {
	return Netconfig{}
}
