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

	"github.com/stratumn/go/jsonhttp"
	"github.com/stratumn/go/store/storehttp"
	"github.com/stratumn/goprivate/rethinkstore"
)

const (
	connectAttempts = 12
	connectTimeout  = 10 * time.Second
)

func orStrings(strs ...string) string {
	for _, s := range strs {
		if s != "" {
			return s
		}
	}
	return ""
}

var (
	create   = flag.Bool("create", false, "create tables and indexes then exit")
	drop     = flag.Bool("drop", false, "drop tables and indexes then exit")
	port     = flag.String("port", storehttp.DefaultPort, "server port")
	url      = flag.String("url", orStrings(os.Getenv("RETHINKSTORE_URL"), rethinkstore.DefaultURL), "URL of the RethinkDB database")
	db       = flag.String("db", orStrings(os.Getenv("RETHINKSTORE_DB"), rethinkstore.DefaultDB), "name of the RethinkDB database")
	hard     = flag.Bool("hard", rethinkstore.DefaultHard, "whether to use hard durability")
	certFile = flag.String("tlscert", "", "TLS certificate file")
	keyFile  = flag.String("tlskey", "", "TLS private key file")
	verbose  = flag.Bool("verbose", storehttp.DefaultVerbose, "verbose output")
	version  = "0.1.0"
	commit   = "00000000000000000000000000000000"
)

func init() {
	log.SetPrefix("rethinkstore ")
}

func main() {
	flag.Parse()

	log.Printf("%s v%s@%s", rethinkstore.Description, version, commit[:7])
	log.Print("Copyright (c) 2016 Stratumn SAS")
	log.Print("All Rights Reserved")
	log.Printf("Runtime %s %s %s", runtime.Version(), runtime.GOOS, runtime.GOARCH)

	a, err := rethinkstore.New(&rethinkstore.Config{
		URL:     *url,
		DB:      *db,
		Hard:    *hard,
		Version: version,
		Commit:  commit,
	})
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

	exists := false
	for i := 1; i <= connectAttempts; i++ {
		if err != nil {
			time.Sleep(connectTimeout)
		}
		if exists, err = a.Exists(); err != nil {
			log.Printf("Unable to connect to %q after %d of %d attempts, retrying in %v", *url, i, connectAttempts, connectTimeout)
		} else {
			if !exists {
				if err = a.Create(); err != nil {
					log.Fatalf("err: %s", err)
				}
				log.Print("Created tables and indexes")
			}
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
