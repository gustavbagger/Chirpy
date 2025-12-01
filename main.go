package main

import (
	"fmt"
	"net/http"
)

func main() {
	servemux := http.NewServeMux()
	server := http.Server{
		Handler: servemux,
		Addr:    ":8080",
	}
	err := server.ListenAndServe()
	if err != nil {
		fmt.Errorf("error: Could not serve server")
		return
	}
}
