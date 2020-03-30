package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
)

func validate(ifAddress, remoteNetwork, remoteGateway string) error {
	// IP address of the local tun interface
	if net.ParseIP(ifAddress) == nil {
		return fmt.Errorf("Invalid Interface IP address")
	}

	// Remote network via the remote tunnel
	if remoteNetwork != "" {
		_, ipNet, err := net.ParseCIDR(remoteNetwork)
		if err != nil {
			return err
		}
		remoteNetwork = ipNet.Network()
	}

	// Remote gateway via the remote tunnel
	if len(remoteGateway) > 0 && net.ParseIP(remoteGateway) == nil {
		return fmt.Errorf("Invalid Remote Gateway IP address")
	}
	return nil
}

func main() {

	var remoteNetwork, remoteGateway, ifAddress string
	connectCmd := flag.NewFlagSet("connect", flag.ExitOnError)
	remoteAddress := connectCmd.String("dst-host", "", "remote host address")
	remotePort := connectCmd.Int("dst-port", 0, "specify the local port to be used")
	connectCmd.StringVar(&ifAddress, "if-address", "192.168.166.1", "Local interface address")
	connectCmd.StringVar(&remoteNetwork, "remote-network", "", "Remote network via the tunnel")
	connectCmd.StringVar(&remoteGateway, "remote-gateway", "", "Remote gateway via the tunnel")

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

	if err := validate(ifAddress, remoteNetwork, remoteGateway); err != nil {
		log.Fatalf("Validation error %v", err)
		os.Exit(1)
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
		// Send configuretion to the server
		text := fmt.Sprintf("remoteNetwork:%s", remoteNetwork)
		// send to socket
		conn.Write([]byte(text + "\n"))
		// listen for reply
		message, _ := bufio.NewReader(conn).ReadString('\n')
		if strings.TrimSpace(message) != text {
			log.Fatalf("Connection error, Sent: %s Received: %s", text, message)
		}
		// Send configuretion to the server
		text = fmt.Sprintf("remoteGateway:%s", remoteGateway)
		// send to socket
		conn.Write([]byte(text + "\n"))
		// listen for reply
		message, _ = bufio.NewReader(conn).ReadString('\n')
		if strings.TrimSpace(message) != text {
			log.Fatalf("Connection error, Sent: %s Received: %s", text, message)
		}
		// Create the tunnel in client mode
		tun, err := NewTunnel(conn, ifAddress, remoteNetwork, "", false)
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
			// Receive the configuration parameters
			ifAddress = "192.168.166.1"
			// will listen for message to process ending in newline (\n)
			message, _ := bufio.NewReader(conn).ReadString('\n')
			// output message received
			fmt.Printf("Message Received: %s", message)
			// process for string received
			m := strings.Split(message, ":")
			if m[0] != "remoteNetwork" {
				log.Fatalf("Connection error, Received: %s Expected: remoteNetwork", m[0])
			}
			remoteNetwork = strings.TrimSpace(m[1])
			// send new string back to client
			conn.Write([]byte(message + "\n"))
			// will listen for message to process ending in newline (\n)
			message, _ = bufio.NewReader(conn).ReadString('\n')
			// output message received
			fmt.Printf("Message Received: %s", message)
			// process for string received
			m = strings.Split(message, ":")
			if m[0] != "remoteGateway" {
				log.Fatalf("Connection error, Received: %s Expected: remoteGateway", m[0])
			}
			remoteGateway = strings.TrimSpace(m[1])
			// send new string back to client
			conn.Write([]byte(message + "\n"))

			// Create the tunnel in server mode
			tun, err := NewTunnel(conn, ifAddress, remoteNetwork, remoteGateway, true)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println("Running the tunnel")
			go func() {
				tun.Run()
				tun.Stop()
			}()
		}
	}

}
