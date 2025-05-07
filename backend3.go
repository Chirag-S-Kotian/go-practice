package main

import (
	"fmt"
	"net/http"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Response from Backend 3")
	})
	fmt.Println("Backend 3 running on port 8083")
	http.ListenAndServe(":8083", nil)
}