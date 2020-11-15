build: $(wildcard *.go) proxy.go
	go build -ldflags "-s -w" -o ssh-agent-proxy

proxy.go: $(wildcard proxy/*.go)
	go generate

clean:
	-rm -f proxy.go proxy.exe ssh-agent-proxy
