package main

import (
	"crypto/sha512"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"math/rand"
	//"strings"
	"errors"
	"sync"
	"time"
)

type PasswordHash []byte

func HashPassword(pw string) (PasswordHash, error) {
	hash := sha512.New()
	_, err := hash.Write([]byte(pw))
	if err != nil {
		return nil, err
	}
	return hash.Sum(nil), nil
}

func (hash PasswordHash) Base64() string {
	return base64.StdEncoding.EncodeToString(hash)
}

// A 128 bit identifier
type HashId [16]byte

// Generate a unique id based on the given time and a 64bit random number
func NewHashId(t time.Time) *HashId {
	id := new(HashId)
	binary.BigEndian.PutUint64(id[:8], uint64(t.UnixNano()))
	binary.BigEndian.PutUint64(id[8:], rand.Uint64())
	return id
}

func HashIdFromString(str string) (*HashId, error) {
	//str = strings.Replace(str, "-", "", -1)
	idBytes, err := hex.DecodeString(str)
	if err != nil {
		return nil, err
	}
	if len(idBytes) > 16 {
		return nil, errors.New("HashId is too long")
	}
	if len(idBytes) < 16 {
		idBytes = append(make([]byte, 16-len(idBytes)), idBytes...)
	}
	id := &HashId{}
	copy(id.Bytes(), idBytes)
	return id, nil
}

func (id *HashId) Bytes() []byte {
	return (*id)[:]
}

// Timestamp component - a UnixNano timestamp
func (id *HashId) Timestamp() int64 {
	if len(id) < 8 {
		panic("HashId is not correct size")
	}
	return int64(binary.BigEndian.Uint64(id[:8]))
}

// Random component
func (id *HashId) Random() int64 {
	return int64(binary.BigEndian.Uint64(id[8:]))
}

func (id *HashId) String() string {
	str := hex.EncodeToString(id.Bytes())
	return str
}

type HashStore struct {
	hashMap   map[HashId]PasswordHash
	totalTime int
	mutex     sync.Mutex
}

func (hs *HashStore) Insert(id *HashId, ph PasswordHash) {
	mutex.Lock()
	defer mutex.Unlock()
	hashMap[*id] = ph
	hs.totalTime += (time.Now.UTC().UnixNano() - id.Timestamp()) / 1000
}
