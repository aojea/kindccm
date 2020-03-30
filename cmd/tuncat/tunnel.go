package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"runtime"

	"github.com/songgao/water"
	"golang.org/x/sync/errgroup"
)

const defaultInterface = "eth0"

// HostInterface represents the TUN interface and the networking configuration
type HostInterface struct {
	ifce   *water.Interface
	netCfg Netconfig
}

// NewHostInterface returns a new HostInterface
func NewHostInterface(ifAddress, remoteNetwork, remoteGateway string, serverMode bool) (HostInterface, error) {
	// Create TUN interface
	// TODO: Windows have some network specific parameters
	// https://github.com/songgao/water/blob/master/params_windows.go
	ifce, err := water.New(water.Config{
		DeviceType: water.TUN,
	})
	if err != nil {
		return HostInterface{}, err
	}
	log.Printf("Interface Name: %s\n", ifce.Name())

	netCfg := NewNetconfig(ifAddress, remoteNetwork, ifce.Name())
	// The network configuration is deleted when the interface is destroyed
	if err := netCfg.SetupNetwork(); err != nil {
		return HostInterface{}, err
	}

	log.Printf("Interface Up: %s\n", ifce.Name())
	// Set up routes to remote network
	via := ifce.Name()
	if runtime.GOOS == "linux" {
		via = remoteGateway
	}
	log.Printf("Add route %s via %s\n", netCfg.route, via)
	if err := netCfg.CreateRoutes(via); err != nil {
		return HostInterface{}, err
	}
	// Masquerade traffic in server mode and Linux
	if serverMode && runtime.GOOS == "linux" {
		log.Printf("Add Masquerade on interface %s\n", defaultInterface)
		if err := netCfg.CreateMasquerade(defaultInterface); err != nil {
			return HostInterface{}, err
		}
	}

	return HostInterface{
		ifce:   ifce,
		netCfg: netCfg,
	}, nil

}

// Tunnel consist in a TCP connection and a HostInterface
// with its networking configuration
type Tunnel struct {
	ifce       HostInterface
	conn       net.Conn
	serverMode bool
}

// NewTunnel create a new Tunnel
func NewTunnel(conn net.Conn, ifAddress, remoteNetwork, remoteGateway string, serverMode bool) (*Tunnel, error) {
	fmt.Println("Create Host Interface ...")
	ifce, err := NewHostInterface(ifAddress, remoteNetwork, remoteGateway, serverMode)
	if err != nil {
		return nil, err
	}

	return &Tunnel{
		ifce:       ifce,
		conn:       conn,
		serverMode: serverMode,
	}, nil
}

// Run the Tunnel copies the data from the conn to the interface
// and viceversa
func (t *Tunnel) Run() {
	var g errgroup.Group

	// Copy from the Tun interface to the connection
	g.Go(func() error {
		for {
			_, err := io.Copy(t.conn, t.ifce.ifce)
			if err != nil {
				return err
			}
		}
	})

	g.Go(func() error {
		for {
			_, err := io.Copy(t.ifce.ifce, t.conn)
			if err != nil {
				return err
			}
		}
	})

	if err := g.Wait(); err != nil {
		log.Println(err)
	}

}

// Stop cleans the routes and closes the connection and the TUN interface
func (t *Tunnel) Stop() {
	// Set up routes to remote network
	dev := t.ifce.ifce.Name()
	if t.serverMode {
		dev = defaultInterface
	}
	t.ifce.netCfg.DeleteRoutes(dev)
	// Masquerade traffic in server mode and Linux
	if t.serverMode && runtime.GOOS == "linux" {
		if err := t.ifce.netCfg.DeleteMasquerade(dev); err != nil {
			log.Println(err)
		}
	}
	t.ifce.ifce.Close()
	t.conn.Close()
}
