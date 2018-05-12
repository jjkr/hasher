package main

import (
	"crypto/sha512"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	//"sync"
	//"time"
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

type Server struct {
	HttpServer http.Server
	InstanceId int // Unique id for this server instance
	HashDelay  int // Delay in milliseconds before displaying hash results
}

func NewServer(port, hashDelayMs int) *Server {
	server := &Server{
		InstanceId: 42,
		HashDelay:  hashDelayMs}
	server.HttpServer.Addr = fmt.Sprintf(":%d", port)
	return server
}

func (server *Server) Close() error {
	return server.HttpServer.Close()
}

func (server *Server) Listen() error {
	return server.HttpServer.ListenAndServe()
}

type Id struct {
	Timestamp  int
	InstanceId int
	Random     int
}

// Generates a unique 64 bit identifier
//func NewIdGenerator() func() uint64 {
//}

// Returns a 32bit unique machine identifier
//func InstanceId() uint {
//id = os.Getenv("HASHER_INSTANCE_ID")
//}

type HashStore struct {
	HashMap map[uint64]string
}

func NewStore() *HashStore {
	store := HashStore{HashMap: make(map[uint64]string)}
	return &store
}

func (store *HashStore) Insert(id uint64, hash string) {
	store.HashMap[id] = hash
}

func (store *HashStore) Find(id uint64) (string, error) {
	return store.HashMap[id], nil
}

func putHash(w http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	err := req.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var id uint64 = 8923892
	io.WriteString(w, strconv.FormatUint(id, 10))
	log.Println(req.Form["password"])
}

func getHash(w http.ResponseWriter, req *http.Request) {
	if len(req.URL.Path) == len("/hash/") {
		putHash(w, req)
		return
	}
	if req.Method != "GET" {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	id, err := strconv.ParseUint(req.URL.Path[len("/hash/"):], 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	fmt.Printf("Got id: %d\n", id)
	io.WriteString(w, strconv.FormatUint(id, 10))

}

func main() {
	//http.HandleFunc("/hash", putHash)
	//http.HandleFunc("/hash/", getHash)
	//log.Fatal(http.ListenAndServe(":8080", nil))
	server := NewServer(8080, 5000)
	server.Listen()
}
