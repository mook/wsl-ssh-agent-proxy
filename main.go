// +build linux

// Package main is the Linux-side listener to present a SSH agent socket.
package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"os/exec"
)

var socketPath = flag.String("socket", "/run/ssh-agent.sock", "Unix socket for SSH agent")

// verbose controls if any output is emitted; by default, the operation is silent.
var verbose = flag.Bool("verbose", false, "Output status on standard error")

func log(format string, args ...interface{}) {
	if *verbose {
		fmt.Fprintf(os.Stderr, format+"\n", args...)
	}
}

func listen(socketPath string) error {
	// Extract the Win32 proxy binary
	proxy, err := extractProxy()
	if err != nil {
		return fmt.Errorf("Error extracting proxy executable: %w", err)
	}
	defer proxy.Close()
	log("Will use proxy at %s\n", proxy)

	// Listen on the Unix socket
	err = removeExistingSocket(socketPath)
	if err != nil {
		return fmt.Errorf("Could not remove existing socket: %w", err)
	}
	addr, err := net.ResolveUnixAddr("unix", socketPath)
	if err != nil {
		return fmt.Errorf("Could not listen on %s: %w", socketPath, err)
	}
	listener, err := net.ListenUnix("unix", addr)
	if err != nil {
		return fmt.Errorf("Could not listen on %s: %w", socketPath, err)
	}
	log("Listening on %s", socketPath)

	// Accept connections
	for {
		conn, err := listener.AcceptUnix()
		if err != nil {
			return fmt.Errorf("Could not accept on %s; %w", socketPath, err)
		}
		go handleConnection(conn, proxy)
	}
}

func handleConnection(conn *net.UnixConn, proxy *proxyFile) error {
	log("Got new connection %+v", conn)

	var args []string
	if *verbose {
		args = append(args, "-verbose")
	}
	cmd := exec.Command(proxy.String(), args...)
	cmd.Stdin = conn
	cmd.Stdout = conn
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		err = fmt.Errorf("Error running command: %s", err)
		fmt.Fprintf(os.Stderr, "%s\n", err)
		return err
	}
	return nil
}

func main() {
	if socketFromEnv, ok := os.LookupEnv("SSH_AUTH_SOCK"); ok {
		*socketPath = socketFromEnv
	}
	flag.Parse()

	err := listen(*socketPath)
	if err != nil {
		log("Error: %s\n", err)
		os.Exit(1)
	}
}
