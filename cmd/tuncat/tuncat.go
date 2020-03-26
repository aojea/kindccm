package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"

	"github.com/songgao/water"
)

func pipeConnIface(conn net.Conn, ifce *water.Interface) {

	// Copy from the connection to the Tun interface
	go func() {
		for {
			_, err := io.Copy(ifce, conn)
			if err != nil {
				// log only the error
				// don't fail if the interface is not ready
				log.Println(err)
			}
		}
	}()

	// Copy from the Tun interface to the connection
	for {
		_, err := io.Copy(conn, ifce)
		if err != nil {
			// return if there is some error
			// the connection is handled out of the loop
			log.Println(err)
			return
		}
	}

}

func main() {
	const defaultNetwork = "192.168.166.1/24"

	connectCmd := flag.NewFlagSet("connect", flag.ExitOnError)
	remoteAddress := connectCmd.String("dst-host", "", "remote host address")
	remotePort := connectCmd.Int("dst-port", 0, "specify the local port to be used")

	listenCmd := flag.NewFlagSet("listen", flag.ExitOnError)
	sourceAddress := listenCmd.String("src-host", "0.0.0.0", "specify the local address to be used (Default: listen in all interfaces)")
	sourcePort := listenCmd.Int("src-port", 0, "specify the local port to be used")

	if len(os.Args) < 2 {
		fmt.Println("usage: tuncat <command> [<args>]")
		fmt.Println("tuncat commands are: ")
		fmt.Println(" connect Connect to a remote host")
		fmt.Println(" listen  Listen on a local port")
		os.Exit(1)
	}

	switch os.Args[1] {

	case "listen":
		listenCmd.Parse(os.Args[2:])
	case "connect":
		connectCmd.Parse(os.Args[2:])
	default:
		fmt.Println("expected 'connect' or 'listen' subcommands")
		os.Exit(1)
	}

	ifce, err := water.New(water.Config{
		DeviceType: water.TUN,
	})
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Interface Name: %s\n", ifce.Name())

	// Connect command
	if connectCmd.Parsed() {
		// Obtain remote port and remote address
		if *remoteAddress == "" || *remotePort == 0 {
			connectCmd.PrintDefaults()
			os.Exit(1)
		}
		// Connect to the remote address
		remoteHost := net.JoinHostPort(*remoteAddress, strconv.Itoa(*remotePort))
		conn, err := net.Dial("tcp", remoteHost)
		if err != nil {
			log.Fatal(err)
		}
		// Copy from the connection to the Tun interface
		pipeConnIface(conn, ifce)
	}

	// Listen command
	if listenCmd.Parsed() {
		if *sourcePort == 0 {
			listenCmd.PrintDefaults()
			os.Exit(1)
		}
		// Listen
		sourceHost := net.JoinHostPort(*sourceAddress, strconv.Itoa(*sourcePort))
		ln, err := net.Listen("tcp", sourceHost)
		if err != nil {
			log.Fatal(err)
		}

		for {
			conn, err := ln.Accept()
			if err != nil {
				log.Fatal(err)
			}
			// only accept one connection
			pipeConnIface(conn, ifce)
		}
	}

}
