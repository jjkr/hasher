package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"sync"
	"time"
)

func logRequest(req *http.Request) {
	log.Printf("%s %s", req.Method, req.URL.Path)
}

type Stats struct {
	Count       int64
	TotalTimeUs int64
}

func (stats *Stats) JsonSummary() ([]byte, error) {
	var average float64
	if stats.Count > 0 {
		average = float64(stats.TotalTimeUs) / float64(stats.Count)
	}
	return json.Marshal(struct {
		Total   int64   `json:"total"`
		Average float64 `json:"average"`
	}{
		stats.Count,
		average,
	})
}

// A Hasher server
type Server struct {
	Done          chan struct{}
	running       bool
	shutdownMutex sync.Mutex
	listener      net.Listener
	httpServer    *http.Server
	hashDelay     time.Duration
	hashMapMutex  sync.Mutex
	hashMap       map[HashId]PasswordHash
	statsMutex    sync.Mutex
	stats         *Stats
}

func StartServer(port int, hashDelay time.Duration) (*Server, error) {
	if port < 0 {
		return nil, errors.New("Port cannot be negative")
	}
	mux := http.NewServeMux()
	bindAddr := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", bindAddr)
	if err != nil {
		return nil, err
	}
	server := &Server{
		Done:     make(chan struct{}),
		running:  true,
		listener: listener,
		httpServer: &http.Server{
			Handler: mux,
		},
		hashDelay: hashDelay,
		hashMap:   make(map[HashId]PasswordHash),
		stats:     &Stats{},
	}
	mux.HandleFunc("/hash", func(w http.ResponseWriter, req *http.Request) {
		//logRequest(req)
		server.PutHash(w, req)
	})
	mux.HandleFunc("/hash/", func(w http.ResponseWriter, req *http.Request) {
		//logRequest(req)
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
	go server.httpServer.Serve(listener) // errors handled in Shutdown
	return server, nil
}

// Gracefully shutdown the server.
// The Done channel will be closed when the server shuts down
func (server *Server) Shutdown() {
	server.shutdownMutex.Lock()
	if server.running {
		server.running = false
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		go func() {
			defer server.shutdownMutex.Unlock()
			defer cancel()
			if err := server.httpServer.Shutdown(ctx); err != nil {
				log.Printf("Server err: %v", err)
				server.httpServer.Close()
			}
			// Explicitly close the listener. If Shutdown is called before the
			// Serve goroutine has run, calling Shutdown on the http.Server will
			// not close the listener, so do it here
			server.listener.Close()
			close(server.Done)
		}()
	}
}

// PUT /hash
func (server *Server) PutHash(w http.ResponseWriter, req *http.Request) {
	startTime := time.Now().UTC()

	defer req.Body.Close()
	if req.Method != "POST" {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	password := req.FormValue("password")
	if len(password) == 0 {
		http.Error(w, "Missing password field", http.StatusBadRequest)
		return
	}

	hashId := NewHashId(startTime)

	// Run long running hash function on a separate goroutine
	go func() {
		time.Sleep(server.hashDelay)
		hash, err := HashPassword(password)
		if err != nil {
			log.Printf("Hash password error: %v\n", err)
			return
		}
		server.hashMapMutex.Lock()
		defer server.hashMapMutex.Unlock()
		server.hashMap[*hashId] = hash
	}()

	w.WriteHeader(http.StatusAccepted)
	io.WriteString(w, hashId.String())

	endTime := time.Now().UTC()
	server.statsMutex.Lock()
	defer server.statsMutex.Unlock()
	server.stats.Count += 1
	server.stats.TotalTimeUs += int64(endTime.Sub(startTime)) / 1000
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

	server.hashMapMutex.Lock()
	defer server.hashMapMutex.Unlock()
	hash := server.hashMap[*id]
	if hash == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	w.Write(hash.Bytes())
}

// GET /stats
func (server *Server) GetStats(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()

	statsJson, err := server.stats.JsonSummary()
	if err != nil {
		log.Printf("Marshal stats json error: %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(statsJson)
}
