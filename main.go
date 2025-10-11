// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"embed"
	"flag"
	"io/fs"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"

	"golang.org/x/tools/present"
)

var (
	httpAddr       = flag.String("http", "127.0.0.1:3999", "HTTP service address (e.g., '127.0.0.1:3999')")
	originHost     = flag.String("orighost", "", "host component of web origin URL (e.g., 'localhost')")
	basePath       = flag.String("base", "", "base path for slide template and static resources")
	contentPath    = flag.String("content", ".", "base path for presentation content")
	generateStatic = flag.String("static", "", "if set, generate static content and exit")
)

//go:embed static templates
var embedFS embed.FS

var (
	fsys   fs.FS = embedFS
	origin *url.URL
)

func main() {
	flag.BoolVar(&present.NotesEnabled, "notes", false, "enable presenter notes (press 'N' from the browser to display them)")
	flag.Parse()

	if *basePath != "" {
		fsys = os.DirFS(*basePath)
	}
	err := initTemplates(fsys)
	if err != nil {
		log.Fatalf("Failed to parse templates: %v", err)
	}

	ln, err := net.Listen("tcp", *httpAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer ln.Close()

	_, port, err := net.SplitHostPort(ln.Addr().String())
	if err != nil {
		log.Fatal(err)
	}

	origin = &url.URL{Scheme: "http"}
	if *originHost != "" {
		if after, ok := strings.CutPrefix(*originHost, "https://"); ok {
			*originHost = after
			origin.Scheme = "https"
		}
		*originHost = strings.TrimPrefix(*originHost, "http://")
		origin.Host = net.JoinHostPort(*originHost, port)
	} else if ln.Addr().(*net.TCPAddr).IP.IsUnspecified() {
		name, _ := os.Hostname()
		origin.Host = net.JoinHostPort(name, port)
	} else {
		reqHost, reqPort, err := net.SplitHostPort(*httpAddr)
		if err != nil {
			log.Fatal(err)
		}
		if reqPort == "0" {
			origin.Host = net.JoinHostPort(reqHost, port)
		} else {
			origin.Host = *httpAddr
		}
	}

	http.Handle("/static/", http.FileServer(http.FS(fsys)))

	log.Printf("Open your web browser and visit %s", origin.String())
	if present.NotesEnabled {
		log.Println("Notes are enabled, press 'N' from the browser to display them.")
	}

	if *generateStatic != "" {
		go func() {
			log.Fatal(http.Serve(ln, nil))
		}()
		walkStatic()
	} else {
		log.Fatal(http.Serve(ln, nil))
	}
}
