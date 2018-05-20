package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	//"strings"
	"testing"
	"time"
)

func TestServerCreateAndShutdown(t *testing.T) {
	server, err := StartServer(8080, 0)
	if err != nil {
		t.Error(err)
	}

	server.Shutdown()
	<-server.Done
}

func TestNewServerPortAlreadyBound(t *testing.T) {
	server, err := StartServer(8080, 0)
	if err != nil {
		t.Error(err)
	}
	duplicateServer, err := StartServer(8080, 0)
	if err == nil {
		t.Error("Expected error to be non-nil")
	}
	if duplicateServer != nil {
		t.Error("Expected duplicateServer to be nil")
	}

	server.Shutdown()
	<-server.Done
}

func TestNewServerNegativePort(t *testing.T) {
	server, err := StartServer(-26, 0)
	if err == nil {
		t.Error("Expected error to be non-nil")
	}
	if server != nil {
		t.Error("Expected server to be nil")
	}
}

func TestHashPost(t *testing.T) {
	server, err := StartServer(8081, time.Nanosecond)
	if err != nil {
		t.Error(err)
	}
	client := &http.Client{}

	postResponse, err := client.PostForm(
		"http://localhost:8081/hash",
		url.Values{"password": {"PeachPie"}})
	if err != nil {
		t.Error(err)
	}
	id, err := ioutil.ReadAll(postResponse.Body)
	postResponse.Body.Close()
	if err != nil {
		t.Error(err)
	}
	if expected := http.StatusAccepted; postResponse.StatusCode != expected {
		t.Errorf("POST /hash got bad status: %d, expected %d\n",
			postResponse.StatusCode, expected)
	}

	_, err = HashIdFromString(string(id))

	server.Shutdown()
	<-server.Done
}

func TestHashPostMissingPassword(t *testing.T) {
	server, err := StartServer(8082, time.Nanosecond)
	if err != nil {
		t.Error(err)
	}
	client := &http.Client{}

	postResponse, err := client.PostForm(
		"http://localhost:8082/hash", url.Values{})
	if err != nil {
		t.Error(err)
	}
	if expectedStatus := http.StatusBadRequest; postResponse.StatusCode != expectedStatus {
		t.Errorf(
			"POST /hash with no password got status: %d, expected: %d",
			postResponse.StatusCode, expectedStatus)
	}

	server.Shutdown()
	<-server.Done
}

func TestHashGetWrongUrl(t *testing.T) {
	server, err := StartServer(8083, time.Nanosecond)
	if err != nil {
		t.Error(err)
	}
	client := &http.Client{}

	getResponse, err := client.Get("http://localhost:8083/hash")
	if err != nil {
		t.Error(err)
	}
	if expectedStatus := http.StatusNotFound; getResponse.StatusCode != expectedStatus {
		t.Errorf(
			"GET /hash got status: %d, expected: %d",
			getResponse.StatusCode, expectedStatus)
	}

	server.Shutdown()
	<-server.Done
}

func TestHashIntegration(t *testing.T) {
	server, err := StartServer(8080, 0)
	if err != nil {
		t.Error(err)
	}
	client := &http.Client{}

	postResponse, err := client.PostForm(
		"http://localhost:8080/hash",
		url.Values{"password": {"PeachPie"}})
	if err != nil {
		t.Error(err)
	}
	id, err := ioutil.ReadAll(postResponse.Body)
	postResponse.Body.Close()
	if err != nil {
		t.Error(err)
	}
	if postResponse.StatusCode != 202 {
		t.Errorf("POST /hash got bad status: %d\n", postResponse.StatusCode)
	}

	getResponse, err := client.Get(
		fmt.Sprintf("http://localhost:8080/hash/%s", id))
	if err != nil {
		t.Error(err)
	}
	hash, err := ioutil.ReadAll(getResponse.Body)
	getResponse.Body.Close()
	if err != nil {
		t.Error(err)
	}
	log.Println(string(hash))

	shutdownResponse, err := client.Get("http://localhost:8080/shutdown")
	if err != nil {
		t.Error(err)
	}
	io.Copy(ioutil.Discard, shutdownResponse.Body)
	shutdownResponse.Body.Close()

	<-server.Done
}

func BenchmarkPostHash(b *testing.B) {
	server, err := StartServer(8082, 5000)
	if err != nil {
		b.Error(err)
	}
	client := &http.Client{}
	formValues := url.Values{"password": {"PeachPie"}}
	for i := 0; i < b.N; i++ {
		resp, err := client.PostForm("http://localhost:8082/hash", formValues)
		if err != nil {
			b.Error(err)
		}
		io.Copy(ioutil.Discard, resp.Body)
		resp.Body.Close()
	}

	server.Shutdown()
	<-server.Done
}
