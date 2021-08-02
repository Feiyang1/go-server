package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const DEFAULT_DIR = "./"

func main() {

	var port = flag.String("p", "8080", "port to start the server")

	flag.Parse()
	positionalArgs := flag.Args()

	directory := DEFAULT_DIR
	if len(positionalArgs) > 0 {
		directory = positionalArgs[0]
	}

	http.HandleFunc("/", makeFileHandler(directory))

	fmt.Printf("Starting server at port %s\n", *port)

	if err := http.ListenAndServe(":"+*port, nil); err != nil {
		log.Fatal(err)
	}
}

func makeFileHandler(directory string) http.HandlerFunc {
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
				dat, err := readFile(filePath)

				if err == nil {
					fmt.Fprintf(w, dat)
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

func readFile(path string) (string, error) {
	dat, err := ioutil.ReadFile(path)

	if err != nil {
		return "", err
	} else {
		return string(dat), err
	}
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
