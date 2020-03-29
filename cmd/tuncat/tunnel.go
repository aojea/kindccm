package main

import (
	"io"
	"log"
	"net"

	"github.com/songgao/water"
)

// HostInterface represents the TUN interface and the networking configuration
type HostInterface struct {
	ifce   *water.Interface
	netCfg Netconfig
}

// NewHostInterface returns a new HostInterface
func NewHostInterface(ifAddress, remoteNetwork string) (HostInterface, error) {
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

	// Create Network configuration
	netCfg := NewNetconfig(ifAddress, remoteNetwork, ifce.Name())
	// The network configuration is deleted when the interface is destroyed
	if err := netCfg.SetupNetwork(); err != nil {
		return HostInterface{}, err
	}
	// Set up routes to remote network
	if err := netCfg.CreateRoutes(); err != nil {
		return HostInterface{}, err
	}

	return HostInterface{
		ifce:   ifce,
		netCfg: netCfg,
	}, nil

}

// Tunnel consist in a TCP connection and a HostInterface
// with its networking configuration
type Tunnel struct {
	ifce HostInterface
	conn net.Conn
}

// NewTunnel create a new Tunnel
func NewTunnel(conn net.Conn, ifAddress, remoteNetwork string) (*Tunnel, error) {
	ifce, err := NewHostInterface(ifAddress, remoteNetwork)
	if err != nil {
		return nil, err
	}

	return &Tunnel{
		ifce: ifce,
		conn: conn,
	}, nil
}

// Run the Tunnel copies the data from the conn to the interface
// and viceversa
func (t *Tunnel) Run() {

	// Copy from the Tun interface to the connection
	go func() {
		for {
			_, err := io.Copy(t.conn, t.ifce.ifce)
			if err != nil {
				// return if there is some error
				// the connection is handled out of the loop
				log.Println(err)
				return
			}
		}
	}()

	for {
		_, err := io.Copy(t.ifce.ifce, t.conn)
		if err != nil {
			// log only the error
			// don't fail if the interface is not ready
			log.Println(err)
			return
		}
	}

}

// Stop cleans the routes and closes the connection and the TUN interface
func (t *Tunnel) Stop() {
	t.ifce.netCfg.DeleteRoutes()
	t.ifce.ifce.Close()
	t.conn.Close()
}
