package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

type GitServer struct {
	repoName string
}

func NewGitServer() *GitServer {
	return &GitServer{
		repoName: "monorepo",
	}
}

// Git HTTP protocol handlers
func (gs *GitServer) handleInfoRefs(w http.ResponseWriter, r *http.Request) {
	service := r.URL.Query().Get("service")

	if service == "git-upload-pack" {
		w.Header().Set("Content-Type", fmt.Sprintf("application/x-%s-advertisement", service))
		w.Header().Set("Cache-Control", "no-cache")

		// Git protocol pkt-line format
		fmt.Fprintf(w, "001e# service=%s\n", service)
		fmt.Fprint(w, "0000")

		// TODO: Generate proper git refs from monorepo state
		// For now, return minimal refs
		fmt.Fprint(w, "003f0000000000000000000000000000000000000000 refs/heads/main\x00multi_ack thin-pack\n")
		fmt.Fprint(w, "0000")
	} else {
		http.Error(w, "Service not supported", http.StatusForbidden)
	}
}

func (gs *GitServer) handleUploadPack(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/x-git-upload-pack-result")
	w.Header().Set("Cache-Control", "no-cache")

	// TODO: Implement proper pack generation from monorepo
	// This would involve:
	// 1. Parsing the want/have refs from client request body
	// 2. Fetching relevant data from poon-server
	// 3. Generating git pack format with proper objects

	// For now, return empty pack
	fmt.Fprint(w, "0008NAK\n")

	// Empty pack file header
	packHeader := []byte{
		'P', 'A', 'C', 'K', // signature
		0, 0, 0, 2, // version 2
		0, 0, 0, 0, // number of objects (0)
	}
	w.Write(packHeader)

	// Pack checksum (20 bytes of zeros for empty pack)
	checksum := make([]byte, 20)
	w.Write(checksum)
}

func (gs *GitServer) setupRoutes() *http.ServeMux {
	mux := http.NewServeMux()

	// Git HTTP protocol endpoints only
	mux.HandleFunc("/info/refs", gs.handleInfoRefs)
	mux.HandleFunc("/git-upload-pack", gs.handleUploadPack)

	// Health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "OK")
	})

	return mux
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	gitServer := NewGitServer()
	mux := gitServer.setupRoutes()

	log.Printf("Poon Git server listening on port %s", port)
	log.Printf("Serving Git HTTP protocol endpoints only")

	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
