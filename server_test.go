package main

import (
	"log"
	"testing"
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

func TestHashStore(t *testing.T) {
	store := NewStore()
	const testKey uint64 = 2
	const testHash string = "xVChQ1m2AB9Yg5AUL"
	store.Insert(testKey, testHash)

	log.Println(store.Find(testKey))
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

func TestServer(t *testing.T) {
	//server := NewServer(8080, 5000)
}
