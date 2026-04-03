package main

import (
	"log"
	"net/http"

	handler "dodgers-win/api"
)

func main() {
	http.HandleFunc("/api/dodgers_win_alerter", handler.Alerter)
	log.Println("Listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
