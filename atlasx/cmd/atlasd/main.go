package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

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

	if err := fs.Parse(args); err != nil {
		return err
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
