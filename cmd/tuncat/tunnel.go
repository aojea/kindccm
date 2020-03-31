package main

import (
	"io"
	"log"
	"net"

	"golang.org/x/sync/errgroup"
)

// Tunnel consist in a TCP connection and a HostInterface
// with its networking configuration
type Tunnel struct {
	ifce HostInterface
	conn net.Conn
}

// NewTunnel create a new Tunnel
func NewTunnel(conn net.Conn, ifce HostInterface) *Tunnel {
	log.Println("Creating Tunnel ...")

	return &Tunnel{
		ifce: ifce,
		conn: conn,
	}
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

	// Copy from the the connection to the Tun interface
	g.Go(func() error {
		for {
			_, err := io.Copy(t.ifce.ifce, t.conn)
			if err != nil {
				return err
			}
		}
	})

	// Don't fail just log it
	if err := g.Wait(); err != nil {
		log.Println(err)
	}

}

// Stop cleans the routes and closes the connection and the TUN interface
func (t *Tunnel) Stop() {
	t.ifce.Delete()
	t.conn.Close()
}
