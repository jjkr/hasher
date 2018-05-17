package main

import (
	"encoding/hex"
	//"log"
	//"strconv"
	"testing"
	"time"
)

func TestHashPassword(t *testing.T) {
	hash, err := HashPassword("GoBroncos!漂亮的马")
	if err != nil {
		t.Error(err)
	}
	// Base64 Encoded sha512 sum
	// echo -n "GoBroncos!漂亮的马" | sha512sum |
	//   awk '{print $1}' | xxd -r -p | base64 -w 0
	expected := "1MCECi+W8NtlU3VATB4O8m3Ppko6h4Dw+gB1B8Tg74cNVnhT32wZoXcks4mKXHYRIRwGpeC6tjNgZiJZrhgCMg=="
	if h := hash.Base64(); h != expected {
		t.Errorf("Wrong hash, got %s", h)
	}
}

func TestBase64EmptyPassword(t *testing.T) {
	hash, err := HashPassword("")
	if err != nil {
		t.Error(err)
	}
	expected := "z4PhNX7vuL3xVChQ1m2AB9Yg5AULVxXcg/SpIdNs6c5H0NE8XYXysP+DGNKHfuwvY7kxvUdBeoGlODJ6+SfaPg=="
	if hash.Base64() != expected {
		t.Errorf("Wrong hash")
	}
}

func TestNewHashId(t *testing.T) {
	NewHashId(time.Now())
}

func TestHashIdTimestamp(t *testing.T) {
	testTime := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	id := NewHashId(testTime)
	if testTime.UnixNano() != id.Timestamp() {
		t.Error("Id timestamp is not correct")
	}
}

func TestHashIdFromString(t *testing.T) {
	s := "152ef6f27b4cf17d365a858149c6e2d1"
	id, err := HashIdFromString(s)
	if err != nil {
		t.Error(err)
	}
	if id.String() != s {
		t.Error("Hash string does not match")
	}
}

func TestHashIdFromStringShort(t *testing.T) {
	x := []byte{12}
	id, err := HashIdFromString(hex.EncodeToString(x))
	if err != nil {
		t.Error(err)
	}
	if len(id) != 16 {
		t.Errorf("Expected length of 16, got %d", len(id))
	}
	if id[len(id)-1] != x[0] {
		t.Errorf("Expected id to equal 0x%x, but was 0x%v", x[0], id)
	}
}

func TestHashIdFromStringNotHex(t *testing.T) {
	_, err := HashIdFromString("foobarxxx")
	if err == nil {
		t.Error("Expected non hex string to fail")
	}
}

func TestHashIdFromStringTooLong(t *testing.T) {
	_, err := HashIdFromString(
		"012345678901234567890123456789012345678901234567890123456789")
	if err == nil {
		t.Error("Expected long string to fail")
	}
}

func TestHashIdStringRoundtrip(t *testing.T) {
	id := NewHashId(time.Now())

	s := id.String()
	otherId, err := HashIdFromString(s)
	if err != nil {
		t.Error(err)
	}
	if s != otherId.String() {
		t.Error("Hash string does not match")
	}
}

func TestHashIdStringPad(t *testing.T) {
	id := NewHashId(time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC))

	if id.Timestamp() != 0 {
		t.Error("Id timestamp is not unix epoch based")
	}
	if len(id.String()) != 32 {
		t.Error("String did not pad")
	}
}

func TestHashStore(t *testing.T) {
	hs := NewHashStore()
	id := NewHashId(time.Now())
	hash, err := HashPassword("password")
	hs.Insert(id, hash)
	if err != nil {
		t.Error(err)
	}
	if got := hs.Get(id); got != hash {
		t.Errorf("Expected %s, got %s", hash.Base64(), got.Base64())
	}
}
