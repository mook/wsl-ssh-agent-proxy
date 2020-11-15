package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
)

// Type proxyFile is a structure describing the temporary proxy executable.
type proxyFile struct {
	ready   bool
	path    string
	process *exec.Cmd
}

func (t *proxyFile) String() string {
	return t.path
}

func (t *proxyFile) Close() error {
	if t.path != "" {
		fmt.Printf("Removing temporary file %s\n", t.path)
		return os.Remove(t.path)
	}
	return nil
}

// extractProxy will extract the Win32 proxy executable, and return a reference
// to it.  If this succeeds, the caller is expected to call Close() on the
// result to remove the temporary file.
func extractProxy() (*proxyFile, error) {
	exeFile, err := ioutil.TempFile("", "ssh-agent-proxy.*.exe")
	if err != nil {
		return nil, fmt.Errorf("Could not create temporary file for proxy: %w", err)
	}
	result := &proxyFile{path: exeFile.Name(), ready: false}
	defer func() {
		exeFile.Close()
		if !result.ready {
			result.Close()
		}
	}()
	err = os.Chmod(exeFile.Name(), 0755)
	if err != nil {
		return nil, fmt.Errorf("Could not make temporary file for proxy executable: %w", err)
	}
	if _, err = exeFile.Write(MustAsset("proxy.exe")); err != nil {
		return nil, fmt.Errorf("Could not write temporary file for proxy: %w", err)
	}
	result.ready = true
	return result, nil
}

//go:generate env GOOS=windows go build "-ldflags=-s -w" -o proxy.exe ./proxy
// Note: go-bindata should be replaced when go1.16 is out with the stdlib
// "embed" package
//go:generate go run github.com/kevinburke/go-bindata/go-bindata -nomemcopy -o proxy.go proxy.exe
