package main

import (
	"embed"
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/Mineru98/disk-viz-viewer/internal/api"
)

//go:embed web/static/*
var staticFS embed.FS

func main() {
	port := flag.Int("port", 8180, "Server port")
	flag.Parse()

	server := api.NewServer(staticFS)
	handler := server.SetupRoutes()

	addr := fmt.Sprintf(":%d", *port)
	fmt.Printf("Starting Disk Usage Viewer server on http://localhost%s\n", addr)

	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatal(err)
	}
}
