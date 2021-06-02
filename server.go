package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
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

	fileServer := http.FileServer(http.Dir(directory))
	http.Handle("/", fileServer)

	fmt.Printf("Starting server at port %s\n", *port)

	if err := http.ListenAndServe(":"+*port, nil); err != nil {
		log.Fatal(err)
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
