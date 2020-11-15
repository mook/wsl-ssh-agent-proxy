// +build windows

package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/Microsoft/go-winio"
)

// pipeName is the name of the named pipe for the SSH agent
var pipeName = flag.String("pipe", "\\\\.\\pipe\\openssh-ssh-agent", "Path to the named pipe")

// verbose controls if any output is emitted; by default, the operation is silent.
var verbose = flag.Bool("verbose", false, "Output status on standard error")

type CloseWriter interface {
	CloseWrite() error
}

func main() {
	flag.Parse()
	pipe, err := winio.DialPipe(*pipeName, nil)
	if err != nil {
		if *verbose {
			fmt.Fprintf(os.Stderr, "Could not open ssh-agent pipe: %s\n", err)
		}
		os.Exit(1)
	}
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		_, err := io.Copy(os.Stdout, pipe)
		if err != nil && *verbose {
			fmt.Fprintf(os.Stderr, "Error reading from ssh agent: %s\n", err)
		}
		wg.Done()
	}()
	go func() {
		_, err := io.Copy(pipe, os.Stdin)
		if err != nil && *verbose {
			fmt.Fprintf(os.Stderr, "Could not write to ssh agent: %s\n", err)
		}
		if messagePipe, ok := pipe.(CloseWriter); ok {
			err = messagePipe.CloseWrite()
			if err != nil && *verbose {
				fmt.Fprintf(os.Stderr, "Could not close pipe to ssh agent: %s\n", err)
			}
		} else {
			if *verbose {
				fmt.Fprintf(os.Stderr, "Warning: pipe is not in message mode\n")
			}
			_, err = pipe.Write(nil)
			if err != nil && *verbose {
				fmt.Fprintf(os.Stderr, "Could not tidy up pipe to ssh agent: %s\n", err)
			}
		}
		err = pipe.Close()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not close SSH agent pipe: %s\n", err)
		}
		wg.Done()
	}()
	if *verbose {
		fmt.Fprintf(os.Stderr, "SSH agent pipe established.\n")
	}
	wg.Wait()
	if *verbose {
		fmt.Fprintf(os.Stderr, "SSH agent proxy terminated.\n")
	}
}
