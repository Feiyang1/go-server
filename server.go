package main

import (
	"bytes"
	"compress/gzip"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/andybalholm/brotli"
)

const DEFAULT_DIR = "./"

type CompressionAlgo int

const (
	GZIP CompressionAlgo = iota
	BROTLI
)

func main() {

	var port = flag.String("p", "8080", "port to start the server")
	var algoArg = flag.String("a", "gzip", "compression algorithm. gzip or brotli")
	flag.Parse()
	positionalArgs := flag.Args()

	directory := DEFAULT_DIR
	if len(positionalArgs) > 0 {
		directory = positionalArgs[0]
	}

	var algo CompressionAlgo
	if strings.ToLower(*algoArg) == "brotli" {
		algo = BROTLI
	} else { // default to GZIP if not brotli
		algo = GZIP
	}
	http.HandleFunc("/", makeFileHandler(directory, algo))

	fmt.Printf("Starting server at port %s\nDirectory: %s\nCompression: %s", *port, directory, *algoArg)

	if err := http.ListenAndServe(":"+*port, nil); err != nil {
		log.Fatal(err)
	}
}

func makeFileHandler(directory string, algo CompressionAlgo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.Error(w, "Method is not supported.", http.StatusNotFound)
			return
		}

		// parse path
		urlParts := strings.Split(r.URL.Path, "/")[1:]
		fileName := urlParts[len(urlParts)-1]
		dirPath := urlParts[0 : len(urlParts)-1]

		filePath := filepath.Join(append([]string{directory}, urlParts[:]...)...)
		// try to load index.html if the requested path is '/'
		if r.URL.Path == "/" {
			filePath = filepath.Join(directory, "index.html")
		}

		fi, err := os.Stat(filePath)

		if err != nil {
			http.Error(w, "File not found", http.StatusNotFound)
		} else {
			// return the content of the requested file
			if !fi.IsDir() {
				dat, err := readAndCompressFile(filePath, algo)

				// TODO: set correct content type for the requested resource
				w.Header().Set("content-type", "text/html")

				if algo == GZIP {
					w.Header().Set("Content-Encoding", "gzip")
				} else {
					w.Header().Set("Content-Encoding", "br")
				}

				if err == nil {
					w.Write(dat)
				} else {
					http.Error(w, "Error reading file", http.StatusInternalServerError)
				}
				return
			} else {
				// TODO: list files in the directory

			}
		}

		_, _, _ = fileName, dirPath, fi
	}
}

func readAndCompressFile(path string, algo CompressionAlgo) ([]byte, error) {
	dat, err := ioutil.ReadFile(path)

	if err != nil {
		return nil, err
	}

	if algo == GZIP {
		return compressWithGzip(dat)
	} else { // Brotli
		return compressWithBrotli(dat)
	}
}

func compressWithGzip(dat []byte) ([]byte, error) {
	var err error = nil
	// gzip content
	var writer bytes.Buffer
	gz := gzip.NewWriter(&writer)

	if _, err = gz.Write(dat); err != nil {
		return nil, err
	}

	if err := gz.Close(); err != nil {
		log.Fatal(err)
	}

	return writer.Bytes(), err
}

func compressWithBrotli(dat []byte) ([]byte, error) {
	var err error = nil
	// brotli content
	var writer bytes.Buffer
	br := brotli.NewWriter(&writer)

	if _, err = br.Write(dat); err != nil {
		return nil, err
	}

	if err := br.Close(); err != nil {
		log.Fatal(err)
	}

	return writer.Bytes(), err
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/hello" {
		http.Error(w, "404 not found", http.StatusNotFound)
		return
	}

	if r.Method != "GET" {
		http.Error(w, "Method is not supported.", http.StatusNotFound)
		return
	}

	fmt.Fprintf(w, "Hello1123!")
}

func formHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		fmt.Fprintf(w, "ParseForm() err: %v", err)
		return
	}

	fmt.Fprintf(w, "POST request successful\n")
	name := r.FormValue("name")
	address := r.FormValue("address")

	fmt.Fprintf(w, "Name = %s\n", name)
	fmt.Fprintf(w, "Address = %s\n", address)
}
