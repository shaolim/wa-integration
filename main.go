package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
)

func main() {
	fmt.Println("hello world")
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	err := http.ListenAndServe(":8080", nil)
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal("Error starting server:", err)
	}
}
