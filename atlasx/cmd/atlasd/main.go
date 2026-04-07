package main

import (
	"errors"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"

	"atlasx/internal/daemon"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "atlasd: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	fs := flag.NewFlagSet("atlasd", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	once := fs.Bool("once", false, "bootstrap once and exit")
	listen := fs.String("listen", daemon.DefaultListenAddr, "listen address")
	allowRemoteControl := fs.Bool("allow-remote-control", false, "allow atlasd to listen on non-loopback addresses")

	if err := fs.Parse(args); err != nil {
		return err
	}
	if !*once {
		if err := validateListenAddress(*listen, *allowRemoteControl); err != nil {
			return err
		}
	}

	status, err := daemon.Bootstrap()
	if err != nil {
		return err
	}
	if *once {
		fmt.Print(status.Render())
		return nil
	}

	mux := daemon.NewMux(status)
	fmt.Printf("atlasd listening on http://%s\n", *listen)
	return http.ListenAndServe(*listen, mux)
}

func validateListenAddress(addr string, allowRemoteControl bool) error {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return fmt.Errorf("invalid listen address %q: %w", addr, err)
	}
	if isLoopbackHost(host) {
		return nil
	}
	if allowRemoteControl {
		return nil
	}
	return errors.New("refusing non-loopback listen address without --allow-remote-control")
}

func isLoopbackHost(host string) bool {
	trimmed := strings.TrimSpace(host)
	if trimmed == "localhost" {
		return true
	}
	ip := net.ParseIP(trimmed)
	return ip != nil && ip.IsLoopback()
}
