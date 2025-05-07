package main

import (
	"fmt"
	"net/http"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Response from Backend 2")
	})
	fmt.Println("Backend 2 running on port 8082")
	http.ListenAndServe(":8082", nil)
}
