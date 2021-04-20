package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
)

func main() {
	// Start the server on given port
	port := 8081
	log.Println("Registring a root endpoint")
	fs := http.FileServer(http.Dir("./static/"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	http.HandleFunc("/", root)
	go http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	log.Printf("Started a server on :%d\n", port)

	// Wait for ^C
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
}
