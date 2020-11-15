package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// removeExistingSocket will delete the socket at the socket path, but only if
// there is nobody currently listening there.
func removeExistingSocket(socketPath string) error {
	// If the socket doesn't exist, do nothing.
	_, err := os.Lstat(socketPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("Could not remove existing SSH auth socket: %w", err)
	}

	// Check for anybody using this socket.
	file, err := os.Open("/proc/net/unix")
	if err != nil {
		return fmt.Errorf("Could not open /proc/net/unix: %w", err)
	}
	scanner := bufio.NewScanner(file)
	if !scanner.Scan() {
		if err = scanner.Err(); err != nil {
			return fmt.Errorf("Could not read /proc/net/unix: %w", err)
		}
		return fmt.Errorf("Could not find header in /proc/net/unix")
	}
	fields := strings.Split(scanner.Text(), " ")
	if len(fields) < 1 || fields[len(fields)-1] != "Path" {
		return fmt.Errorf("Could not find Path in /proc/net/unix, got %v", fields)
	}
	for scanner.Scan() {
		entries := strings.SplitN(scanner.Text(), " ", len(fields))
		if len(entries) < len(fields) {
			// There is no path associated with this socket
			continue
		}
		path := entries[len(entries)-1]
		if path == socketPath {
			return fmt.Errorf("Socket %s is busy", socketPath)
		}
	}
	if err = scanner.Err(); err != nil {
		return fmt.Errorf("Could not read /proc/net/unix: %w", err)
	}

	// If we get here, the socket exists and is safe to delete
	err = os.Remove(socketPath)
	if err != nil {
		return fmt.Errorf("Could not remove socket %s: %w", socketPath, err)
	}

	return nil
}
