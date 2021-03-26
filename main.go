package main

import (
	"fmt"
	"net/http"
)

func main() {
	fmt.Println("Starting server at localhost:8080...\nStop the server by using ctrl^c")
	http.HandleFunc("/", root)
	http.ListenAndServe(":8080", nil)
}
