package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// removeExistingSocket will delete the socket at the socket path, but only if
// there is nobody currently listening there.  Returns true if we have
// successfully removed the socket, or false if somebody is already listening on
// it (and therefore we should exit).
func removeExistingSocket(socketPath string) (bool, error) {
	// Check for anybody using this socket.
	inUse, err := isSocketInUse(socketPath)
	if err != nil {
		return false, fmt.Errorf("Could not check if socket %s is in use: %w", socketPath, err)
	}
	if inUse {
		log("Socket %s already in use", socketPath)
		return false, nil
	}

	// If we get here, the socket exists and is safe to delete
	err = os.Remove(socketPath)
	if err != nil && !os.IsNotExist(err) {
		return false, fmt.Errorf("Could not remove socket %s: %w", socketPath, err)
	}

	log("Socket %s is available", socketPath)
	return true, nil
}

// isSocketInUse returns true if some program is already listening on the given
// unix socket path.
func isSocketInUse(socketPath string) (bool, error) {
	file, err := os.Open("/proc/net/unix")
	if err != nil {
		return false, fmt.Errorf("Could not open /proc/net/unix: %w", err)
	}
	scanner := bufio.NewScanner(file)
	if !scanner.Scan() {
		if err = scanner.Err(); err != nil {
			return false, fmt.Errorf("Could not read /proc/net/unix: %w", err)
		}
		return false, fmt.Errorf("Could not find header in /proc/net/unix")
	}
	fields := strings.Fields(scanner.Text())
	if len(fields) < 1 || fields[len(fields)-1] != "Path" {
		return false, fmt.Errorf("Could not find Path in /proc/net/unix, got %v", fields)
	}
	for scanner.Scan() {
		entries := strings.SplitN(scanner.Text(), " ", len(fields))
		if len(entries) < len(fields) {
			// There is no path associated with this socket
			continue
		}
		path := entries[len(entries)-1]
		if path == socketPath {
			return true, nil
		}
	}
	if err = scanner.Err(); err != nil {
		return false, fmt.Errorf("Could not read /proc/net/unix: %w", err)
	}
	return false, nil
}
