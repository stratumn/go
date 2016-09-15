// Copyright 2016 Stratumn SAS. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// that can be found in the LICENSE file.

package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/lib/pq"
	"github.com/stratumn/go/jsonhttp"
	"github.com/stratumn/go/store/storehttp"
	"github.com/stratumn/goprivate/postgresstore"
)

const (
	connectAttempts = 12
	connectTimeout  = 10 * time.Second
	noTableCode     = pq.ErrorCode("42P01")
)

var (
	create   = flag.Bool("create", false, "create tables and indexes then exit")
	drop     = flag.Bool("drop", false, "drop tables and indexes then exit")
	port     = flag.String("port", storehttp.DefaultPort, "server port")
	url      = flag.String("url", postgresstore.DefaultURL, "URL of the PostgreSQL database")
	certFile = flag.String("tlscert", "", "TLS certificate file")
	keyFile  = flag.String("tlskey", "", "TLS private key file")
	verbose  = flag.Bool("verbose", storehttp.DefaultVerbose, "verbose output")
	version  = "0.1.0"
	commit   = "00000000000000000000000000000000"
)

func init() {
	log.SetPrefix("postgresstore ")
}

func main() {
	flag.Parse()

	log.Printf("%s v%s@%s", postgresstore.Description, version, commit[:7])
	log.Print("Copyright (c) 2016 Stratumn SAS")
	log.Print("All Rights Reserved")
	log.Printf("Runtime %s %s %s", runtime.Version(), runtime.GOOS, runtime.GOARCH)

	a, err := postgresstore.New(&postgresstore.Config{URL: *url, Version: version, Commit: commit})
	if err != nil {
		log.Fatalf("Fatal: %s", err)
	}

	if *create {
		if err := a.Create(); err != nil {
			log.Fatalf("Fatal: %s", err)
		}
		log.Print("Created tables and indexes")
		os.Exit(0)
	}

	if *drop {
		if err := a.Drop(); err != nil {
			log.Fatalf("Fatal: %s", err)
		}
		log.Print("Dropped tables and indexes")
		os.Exit(0)
	}

	for i := 1; i <= connectAttempts; i++ {
		if err != nil {
			time.Sleep(connectTimeout)
		}
		if err = a.Prepare(); err != nil {
			if e, ok := err.(*pq.Error); ok && e.Code == noTableCode {
				if err = a.Create(); err != nil {
					log.Fatalf("Fatal: %s", err)
				}
				log.Print("Created tables and indexes")
			} else {
				log.Printf("Unable to connect to %q after %d of %d attempts, retrying in %v", *url, i, connectAttempts, connectTimeout)
			}
		} else {
			break
		}
	}
	if err != nil {
		log.Fatalf("Fatal: %s", err)
	}

	go func() {
		sigc := make(chan os.Signal)
		signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM)
		sig := <-sigc
		log.Printf("Got signal %q", sig)
		log.Print("Stopped")
		os.Exit(0)
	}()

	c := &jsonhttp.Config{
		Port:     *port,
		CertFile: *certFile,
		KeyFile:  *keyFile,
		Verbose:  *verbose,
	}
	h := storehttp.New(a, c)

	log.Printf("Listening on %q", *port)
	if err := h.ListenAndServe(); err != nil {
		log.Fatalf("Fatal: %s", err)
	}
}
