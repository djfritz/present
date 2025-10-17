package main

import (
	"bytes"
	"errors"
	"io"
	"io/fs"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

func walkStatic() {
	// create the base directory
	_, err := os.Stat(*generateStatic)
	if !errors.Is(err, fs.ErrNotExist) {
		log.Fatal("static output directory exists")
	}
	err = os.Mkdir(*generateStatic, 0770)
	if err != nil {
		log.Fatal(err)
	}

	// emit the static directory
	fs.WalkDir(fsys, "static", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			log.Fatal(err)
		}
		if path == "." || path == ".." {
			return nil
		}

		if d.IsDir() {
			err = os.Mkdir(filepath.Join(*generateStatic, path), 0777)
			if err != nil {
				log.Fatal(err)
			}
		} else {
			data, err := fs.ReadFile(fsys, path)
			if err != nil {
				log.Fatal(err)
			}
			err = ioutil.WriteFile(filepath.Join(*generateStatic, path), data, 0666)
			if err != nil {
				log.Fatal(err)
			}
		}
		return nil
	})

	// walk the content directory
	root := os.DirFS(*contentPath)
	fs.WalkDir(root, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			log.Fatal(err)
		}
		if path == "." || path == ".." {
			return nil
		}

		if d.IsDir() {
			err = os.Mkdir(filepath.Join(*generateStatic, path), 0777)
			if err != nil {
				log.Fatal(err)
			}
		}

		// instead of copying the file, we grab the output from the present server
		urlpath, err := url.JoinPath(origin.String(), path)
		if err != nil {
			log.Fatal(err)
		}
		resp, err := http.Get(urlpath)
		if err != nil {
			log.Fatal(err)
		}
		var b bytes.Buffer
		io.Copy(&b, resp.Body)
		resp.Body.Close()
		output := b.Bytes()

		var fpath string
		if d.IsDir() {
			fpath = filepath.Join(*generateStatic, path, "index.html")
			output = bytes.ReplaceAll(output, []byte(".slide\">"), []byte(".slide.html\">"))
			output = bytes.ReplaceAll(output, []byte(".article\">"), []byte(".article.html\">"))
		} else {
			fpath = filepath.Join(*generateStatic, path)
			if strings.HasSuffix(fpath, ".slide") || strings.HasSuffix(fpath, ".article") {
				fpath += ".html"
			}
		}
		f, err := os.Create(fpath)
		if err != nil {
			log.Fatal(err)
		}
		_, err = f.Write(output)
		if err != nil {
			log.Fatal(err)
		}
		f.Close()
		return nil
	})

	// get the root index
	urlpath, err := url.JoinPath(origin.String())
	if err != nil {
		log.Fatal(err)
	}
	resp, err := http.Get(urlpath)
	if err != nil {
		log.Fatal(err)
	}
	var b bytes.Buffer
	io.Copy(&b, resp.Body)
	resp.Body.Close()
	output := b.Bytes()
	output = bytes.ReplaceAll(output, []byte(".slide\">"), []byte(".slide.html\">"))
	output = bytes.ReplaceAll(output, []byte(".article\">"), []byte(".article.html\">"))
	resp.Body.Close()

	fpath := filepath.Join(*generateStatic, "index.html")
	f, err := os.Create(fpath)
	if err != nil {
		log.Fatal(err)
	}
	_, err = f.Write(output)
	if err != nil {
		log.Fatal(err)
	}

	f.Close()
}
