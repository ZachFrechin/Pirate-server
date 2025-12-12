package main

import (
	"log"
	"net/http"
	"os"

	"game-server/internal/repository/inmem"
	"game-server/internal/transport/httpapi"
	"game-server/internal/transport/ws"
	"game-server/internal/usecase"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	store := inmem.NewLobbyStore()
	service := usecase.NewLobbyService(store)
	wsServer := ws.NewServer(service)

	mux := http.NewServeMux()
	mux.Handle("/healthz", httpapi.HealthHandler())
	mux.Handle("/ws", wsServer.Handler())

	addr := ":" + port
	log.Printf("listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}
