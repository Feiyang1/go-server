package main

import (
	"bytes"
	"compress/gzip"
	"embed"
	"errors"
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/andybalholm/brotli"
)

//go:embed listDir.template.html
var static embed.FS

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

	http.FileServer(http.FS(static))

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
		sitePath := filepath.Join(urlParts[:]...)

		dat, contentType, err := loadPage(w, filePath, sitePath)

		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			w.Header().Set("content-type", contentType)

			var compressedData []byte
			if algo == GZIP {
				compressedData, err = compressWithGzip(dat)
				w.Header().Set("Content-Encoding", "gzip")
			} else {
				compressedData, err = compressWithBrotli(dat)
				w.Header().Set("Content-Encoding", "br")
			}

			if err == nil {
				w.Write(compressedData)
			} else {
				http.Error(w, "Internal error", http.StatusInternalServerError)
			}
		}

		_, _ = fileName, dirPath
	}
}

/*
* returns the raw data, content type and error
 */
func loadPage(w http.ResponseWriter, filePath string, sitePath string) ([]byte, string, error) {
	fi, err := os.Stat(filePath)

	if err != nil {
		fmt.Printf("error for %s, %s", filePath, err.Error())
		return nil, "", errors.New("invalid path")
	} else {
		// a file is being requests, return its content
		if !fi.IsDir() {
			dat, err := readFile(filePath)
			contentType := getContentType(filePath)
			return dat, contentType, err
		} else { // path is a directory

			// try to load index.html
			indexFilePath := filepath.Join(filePath, "index.html")

			indexFileInfo, indexFileErr := os.Stat(indexFilePath)
			if indexFileErr != nil || indexFileInfo.IsDir() {
				// list the directory if index.html file doesn't exist or isn't a file
				dat, err := listDirectory(filePath, sitePath)
				return dat, "text/html", err
			} else {
				return loadPage(w, indexFilePath, sitePath)
			}
		}
	}
}

type FileInfo struct {
	Path string
	Name string
}

// directory is guaranteed to be a valid path
func listDirectory(directoryPath string, sitePath string) ([]byte, error) {
	files, err := ioutil.ReadDir(directoryPath)
	if err != nil {
		return nil, err
	}

	var items []FileInfo

	tmpl := template.Must(template.ParseFS(static, "listDir.template.html"))
	for _, f := range files {
		items = append(items, FileInfo{Name: f.Name(), Path: sitePath})
	}

	buff := new(bytes.Buffer)
	if err = tmpl.Execute(buff, items); err != nil {
		return nil, err
	}

	return buff.Bytes(), nil
}

func getContentType(filePath string) string {
	parts := strings.Split(filePath, ".")

	// no extension found
	if len(parts) < 2 {
		return "text/plain"
	} else {
		extension := parts[len(parts)-1]

		switch strings.ToLower(extension) {
		case "css":
			return "text/css"
		case "js":
			return "text/javascript"
		case "html":
			return "text/html"
		default: // unknown type, use text/plain
			return "text/plain"
		}
	}
}

func readFile(path string) ([]byte, error) {
	return ioutil.ReadFile(path)
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
