package main

import (
	"fmt"
	"net/http"
)

func throwErr(w http.ResponseWriter, r *http.Request, err error, code int) {
	fmt.Fprintln(w, fmt.Sprintf("Error %d", code))
	fmt.Println(err.Error())
	return
}

func getGETParam(key string, w http.ResponseWriter, r *http.Request) (val string) {
	keys, ok := r.URL.Query()[key]
	if ok && len(keys[0]) >= 1 {
		return keys[0]
	} else {
		val = ""
	}
	return
}
