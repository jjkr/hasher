package main

import (
	"context"
	//"crypto/sha512"
	//"encoding/base64"
	//"encoding/binary"
	//"encoding/hex"
	"encoding/json"
	//"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"runtime"
	//"strconv"
	//"strings"
	//"sync"
	"time"
)

func logRequest(req *http.Request) {
	log.Printf("%s %s", req.Method, req.URL.Path)
}

// A Hasher server
type Server struct {
	Done       chan struct{}
	httpServer *http.Server
	hashDelay  time.Duration
	hashStore  *HashStore
}

func NewServer(port int, hashDelay time.Duration) *Server {
	if port < 0 {
		panic("Port cannot be negative")
	}
	mux := http.NewServeMux()
	server := &Server{
		Done: make(chan struct{}),
		httpServer: &http.Server{
			Addr:    fmt.Sprintf(":%d", port),
			Handler: mux,
		},
		hashDelay: hashDelay,
		hashStore: NewHashStore(),
	}
	mux.HandleFunc("/hash", func(w http.ResponseWriter, req *http.Request) {
		//logRequest(req)
		server.PutHash(w, req)
	})
	mux.HandleFunc("/hash/", func(w http.ResponseWriter, req *http.Request) {
		logRequest(req)
		server.GetHash(w, req)
	})
	mux.HandleFunc("/stats", func(w http.ResponseWriter, req *http.Request) {
		logRequest(req)
		server.GetStats(w, req)
	})
	mux.HandleFunc("/shutdown", func(w http.ResponseWriter, req *http.Request) {
		logRequest(req)
		server.Shutdown()
	})
	mux.HandleFunc("/slow", func(w http.ResponseWriter, req *http.Request) {
		logRequest(req)
		time.Sleep(10 * time.Second)
	})
	go server.httpServer.ListenAndServe() // errors deferred to shutdown
	return server
}

// Gracefully shutdown the server.
// The Done channel will be closed when the server shuts down
func (server *Server) Shutdown() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	go func() {
		defer cancel()
		if err := server.httpServer.Shutdown(ctx); err != nil {
			log.Printf("Server err: %v", err)
		}
		close(server.Done)
	}()
}

// PUT /hash
func (server *Server) PutHash(w http.ResponseWriter, req *http.Request) {
	startTime := time.Now().UTC()

	defer req.Body.Close()
	if req.Method != "POST" {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	err := req.ParseForm()
	if err != nil {
		log.Printf("Failed to parse form: %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	passwordForm := req.Form["password"]
	if len(passwordForm) < 1 {
		http.Error(w, "Missing password field", http.StatusBadRequest)
		return
	}

	hashId := NewHashId(startTime)

	hash, err := HashPassword(passwordForm[0])
	passwordForm[0] = "" // zero out password pointer
	if err != nil {
		log.Printf("Hash password error: %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	server.hashStore.Insert(hashId, hash)

	io.WriteString(w, hashId.String())
}

// GET /hash/:id
func (server *Server) GetHash(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()
	baseLen := len("/hash/")
	if len(req.URL.Path) == baseLen {
		server.PutHash(w, req)
		return
	}
	if req.Method != "GET" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	id, err := HashIdFromString(req.URL.Path[baseLen:])
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if id.Timestamp() > (time.Now().UTC().Add(-server.hashDelay)).UnixNano() {
		log.Printf("Hash id %v not available yet\n", id)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	hash := server.hashStore.Get(id)
	if hash == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	io.WriteString(w, hash.Base64())
}

// GET /stats
func (server *Server) GetStats(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()

	statsJson, err := json.Marshal(server.hashStore.Stats())
	if err != nil {
		log.Printf("Marshal stats json error: %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(statsJson)
}

func main() {
	runtime.GOMAXPROCS(16)
	rand.Seed(time.Now().UTC().UnixNano())

	port := 8080
	server := NewServer(port, 5*time.Second)
	log.Printf("Starting hasher server on port %d\n", port)
	<-server.Done
}
