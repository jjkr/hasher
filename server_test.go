package main

import (
	"log"
	"net/http"
	"net/url"
	"testing"
	"time"
)

func TestHashPassword(t *testing.T) {
	pw := "GoBroncos!漂亮的马"
	// Base64 Encoded sha512 sum
	// echo -n "GoBroncos!漂亮的马" | sha512sum |
	//   awk '{print $1}' | xxd -r -p | base64 -w 0
	expected := "1MCECi+W8NtlU3VATB4O8m3Ppko6h4Dw+gB1B8Tg74cNVnhT32wZoXcks4mKXHYRIRwGpeC6tjNgZiJZrhgCMg=="
	h, err := HashPassword(pw)
	if err != nil {
		t.Error(err)
	}
	if h != expected {
		t.Errorf("Wrong hash!")
	}
}

func TestHashPasswordEmpty(t *testing.T) {
	h, err := HashPassword("")
	if err != nil {
		t.Error(err)
	}
	if h != "z4PhNX7vuL3xVChQ1m2AB9Yg5AULVxXcg/SpIdNs6c5H0NE8XYXysP+DGNKHfuwvY7kxvUdBeoGlODJ6+SfaPg==" {
		t.Errorf("Wrong hash!")
	}
}

func BenchmarkHashPassword(b *testing.B) {
	pw := "hello"
	for i := 0; i < b.N; i++ {
		_, err := HashPassword(pw)
		if err != nil {
			b.Error(err)
		}
	}
}

func TestGenerateId(t *testing.T) {
	id1, _ := GenerateId()
	time.Sleep(time.Millisecond)
	id2, _ := GenerateId()
	log.Printf("id1: %s\n", id1)
	log.Printf("id2: %s\n", id2)
	if id2 <= id1 {
		t.Errorf("Id2(%s) is less than Id1(%s)", id2, id1)
	}
}

func BenchmarkGenerateId(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := GenerateId()
		if err != nil {
			b.Error(err)
		}
	}
}

func TestServer(t *testing.T) {
	server := NewServer(8080, time.Millisecond)
	go server.Listen()
	server.Close()
}

func TestPostHash(t *testing.T) {
	server := NewServer(8080, 5000)
	go server.Listen()

	resp, err := http.PostForm("http://localhost:8080/hash", url.Values{"password": {"PeachPie"}})
	if err != nil {
		t.Error(err)
	}
	log.Print(resp)

	if resp.StatusCode != 200 {
		t.Errorf("POST /hash got bad status: %d\n", resp.StatusCode)
	}

	server.Close()
}

func BenchmarkPostHash(b *testing.B) {
	server := NewServer(8080, 5000)
	go server.Listen()
	for i := 0; i < b.N; i++ {
		resp, err := http.PostForm("http://localhost:8080/hash", url.Values{"password": {"PeachPie"}})
		if err != nil {
			b.Error(err)
		}
		defer resp.Body.Close()
	}

	server.Close()
}
