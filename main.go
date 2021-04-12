package main

import (
	"fmt"
	"net/http"
)

func main() {
	fmt.Println("Serwer odpalony pod adresem localhost:8080...\nMożna go zatrzymać używając ctrl^c lub zamykając okno")
	http.HandleFunc("/", root)
	http.ListenAndServe(":8080", nil)
}
