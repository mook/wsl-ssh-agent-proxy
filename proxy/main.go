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

func log(format string, args ...interface{}) {
	if *verbose {
		fmt.Fprintf(os.Stderr, format+"\n", args...)
	}
}

func main() {
	flag.Parse()
	pipe, err := winio.DialPipe(*pipeName, nil)
	if err != nil {
		log("Could not open ssh-agent pipe: %s", err)
		os.Exit(1)
	}
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		_, err := io.Copy(os.Stdout, pipe)
		if err != nil {
			log("Error reading from ssh agent: %s", err)
		}
		wg.Done()
	}()
	go func() {
		_, err := io.Copy(pipe, os.Stdin)
		if err != nil {
			log("Could not write to ssh agent: %s", err)
		}
		if messagePipe, ok := pipe.(CloseWriter); ok {
			err = messagePipe.CloseWrite()
			if err != nil {
				log("Could not close pipe to ssh agent: %s", err)
			}
		} else {
			log("Warning: pipe is not in message mode")
			_, err = pipe.Write(nil)
			if err != nil {
				log("Could not tidy up pipe to ssh agent: %s", err)
			}
		}
		err = pipe.Close()
		if err != nil {
			log("Could not close SSH agent pipe: %s", err)
		}
		wg.Done()
	}()
	log("SSH agent pipe established.")
	wg.Wait()
	log("SSH agent proxy terminated.")
}
