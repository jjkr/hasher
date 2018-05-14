package main

import (
	"context"
	"crypto/sha512"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	//"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"runtime"
	//"strconv"
	"strings"
	"sync"
	"time"
)

// Base64 Encoded sha512 sum of given password
func HashPassword(pw string) (string, error) {
	hash := sha512.New()
	_, err := hash.Write([]byte(pw))
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(hash.Sum(nil)), nil
}

// A 128 bit identifier
type HashId []byte

func NewHashId(t time.Time) HashId {
	id := make([]byte, 16)
	binary.BigEndian.PutUint64(id[:8], uint64(t.UnixNano()))
	binary.BigEndian.PutUint64(id[8:], rand.Uint64())
	return id
}

// Returns nil if str is not a valid hash id
func HashIdFromString(str string) HashId {
	//str = strings.Replace(str, "-", "", -1)
	idBytes, err := hex.DecodeString(str)
	if err != nil {
		log.Printf("HashIdFromString error: %v\n", err)
		return nil
	}
	if len(idBytes) > 16 {
		return nil
	}
	if len(idBytes) < 16 {
		idBytes = append(make([]byte, 16-len(idBytes)), idBytes...)
	}
	return idBytes
}

// Timestamp component - a UnixNano timestamp
func (id HashId) Timestamp() int64 {
	return int64(binary.BigEndian.Uint64(id[:8]))
}

// Random component
func (id HashId) Random() int64 {
	return int64(binary.BigEndian.Uint64(id[8:]))
}

func (id HashId) String() string {
	str := hex.EncodeToString(id)
	// pad with zeros
	if len(str) < 32 {
		str = strings.Repeat("0", 32-len(str)) + str
	}
	return str // fmt.Sprintf("%s-%s-%s-%s-%s", str[:8], str[8:12], str[12:16], str[16:20], str[20:32])
}

type Stats struct {
	Total   int64 `json:"total"`
	Average int64 `json:"average"`
}

type StatsCounter struct {
	Total       int64
	TotalTimeUs int64
}

func (counter *StatsCounter) RecordStat(duration time.Duration) {
	counter.Total += 1
	counter.TotalTimeUs += duration.Nanoseconds() / 1000
}

func (counter *StatsCounter) Add(other *StatsCounter) {
	counter.Total += other.Total
	counter.TotalTimeUs += other.TotalTimeUs
}

func (counter *StatsCounter) Stats() *Stats {
	stats := &Stats{
		Total: counter.Total,
	}
	if counter.Total != 0 {
		stats.Average = counter.TotalTimeUs / counter.Total
	}
	return stats
}

func logRequest(req *http.Request) {
	log.Printf("%s %s", req.Method, req.URL.Path)
}

// A Hasher server
type Server struct {
	httpServer          *http.Server
	HashDelay           time.Duration
	Done                chan struct{}
	hashMap             map[string]string
	hashMapMutex        sync.Mutex
	statsCounters       []*StatsCounter
	statsCounterMutexes []sync.Mutex
}

func NewServer(port int, hashDelay time.Duration) *Server {
	if port < 0 {
		panic("Port cannot be negative")
	}
	mux := http.NewServeMux()
	server := &Server{
		httpServer: &http.Server{
			Addr:    fmt.Sprintf(":%d", port),
			Handler: mux,
		},
		HashDelay:           hashDelay,
		Done:                make(chan struct{}),
		hashMap:             make(map[string]string),
		statsCounters:       make([]*StatsCounter, 256),
		statsCounterMutexes: make([]sync.Mutex, 256),
	}
	for i := range server.statsCounters {
		server.statsCounters[i] = new(StatsCounter)
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
	go server.httpServer.ListenAndServe()
	return server
}

// Gracefully shutdown the server, waiting for existing requests to complete
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
	hashIdString := hashId.String()

	hash, err := HashPassword(passwordForm[0])
	if err != nil {
		log.Printf("Hash password error: %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	passwordForm[0] = ""

	stripe := hashId.Random() & 0xff
	func() {
		server.hashMapMutex.Lock()
		defer server.hashMapMutex.Unlock()
		server.hashMap[hashIdString] = hash
	}()

	io.WriteString(w, hashIdString)

	server.statsCounterMutexes[stripe].Lock()
	defer server.statsCounterMutexes[stripe].Unlock()
	server.statsCounters[stripe].RecordStat(time.Since(startTime))
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

	id := HashIdFromString(req.URL.Path[baseLen:])
	if id == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if id.Timestamp() > (time.Now().UTC().Add(-server.HashDelay)).UnixNano() {
		log.Printf("Hash id %v not available yet\n", id)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	server.hashMapMutex.Lock()
	defer server.hashMapMutex.Unlock()
	hash, ok := server.hashMap[id.String()]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	io.WriteString(w, hash)
}

// GET /stats
func (server *Server) GetStats(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()
	sumCounter := StatsCounter{}
	for i, c := range server.statsCounters {
		server.statsCounterMutexes[i].Lock()
		defer server.statsCounterMutexes[i].Unlock()
		sumCounter.Add(c)
	}
	statsJson, err := json.Marshal(sumCounter.Stats())
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
