package main

import (
	"log"
	"net/http"

	"aftersalecleaner/internal/tickets"
)

func main() {
	handler := tickets.NewHandler(tickets.NewLocalTextCleaner())
	server := &http.Server{Addr: ":8080", Handler: handler}
	log.Printf("after-sale ticket cleaner listening on %s", server.Addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
