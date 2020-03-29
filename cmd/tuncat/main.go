package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
)

func main() {
	ifAddress := flag.String("if-address", "192.168.166.1", "Local interface address")
	remoteNetwork := flag.String("remote-network", "", "Remote network via the tunnel")

	connectCmd := flag.NewFlagSet("connect", flag.ExitOnError)
	remoteAddress := connectCmd.String("dst-host", "", "remote host address")
	remotePort := connectCmd.Int("dst-port", 0, "specify the local port to be used")

	listenCmd := flag.NewFlagSet("listen", flag.ExitOnError)
	sourceAddress := listenCmd.String("src-host", "0.0.0.0", "specify the local address to be used")
	sourcePort := listenCmd.Int("src-port", 0, "specify the local port to be used")

	if len(os.Args) < 2 {
		fmt.Println("usage: tuncat [<args>] <command>")
		flag.PrintDefaults()
		fmt.Println("tuncat commands are: ")
		fmt.Println(" connect [<args>] Connect to a remote host")
		fmt.Println(" listen [<args>] Listen on a local port")
		os.Exit(1)
	}

	switch os.Args[1] {

	case "listen":
		listenCmd.Parse(os.Args[2:])
	case "connect":
		connectCmd.Parse(os.Args[2:])
	default:
		fmt.Println("usage: tuncat [<args>] <command>")
		flag.PrintDefaults()
		fmt.Println("tuncat commands are: ")
		fmt.Println(" connect [<args>] Connect to a remote host")
		connectCmd.PrintDefaults()
		fmt.Println(" listen [<args>] Listen on a local port")
		listenCmd.PrintDefaults()
		os.Exit(1)
	}

	// Global configuration
	flag.Parse()

	// IP address of the tun interface
	if net.ParseIP(*ifAddress) == nil {
		fmt.Errorf("Invalid Interface IP address")
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Remote network via the tun interface
	if *remoteNetwork != "" {
		_, _, err := net.ParseCIDR(*remoteNetwork)
		if err != nil {
			log.Fatal(err)
		}
	}

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
		// Create the tunnel
		tun, err := NewTunnel(conn, *ifAddress, *remoteNetwork)
		if err != nil {
			log.Fatal(err)
		}
		tun.Run()
		tun.Stop()
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
			// Create the tunnel
			tun, err := NewTunnel(conn, *ifAddress, *remoteNetwork)
			if err != nil {
				log.Fatal(err)
			}
			go func() {
				tun.Run()
				tun.Stop()
			}()
		}
	}

}
