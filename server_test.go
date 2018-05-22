package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"
	"time"
)

func TestStatsJson(t *testing.T) {
	var count int64 = 129
	var totalTime int64 = 23723
	stats := Stats{
		Count:       count,
		TotalTimeUs: totalTime,
	}
	jsonBytes, err := stats.JsonSummary()
	if err != nil {
		t.Error(err)
	}
	expectedAvg := float64(totalTime) / float64(count)
	expected := fmt.Sprintf("{\"total\":%d,\"average\":%.14f}",
		count, expectedAvg)
	if string(jsonBytes) != expected {
		t.Errorf("Expected %s, got %s", expected, jsonBytes)
	}
}

func TestStatsZeroJson(t *testing.T) {
	stats := Stats{}
	jsonBytes, err := stats.JsonSummary()
	if err != nil {
		t.Error(err)
	}
	expected := "{\"total\":0,\"average\":0}"
	if string(jsonBytes) != expected {
		t.Errorf("Expected %s, got %s", expected, jsonBytes)
	}
}

func TestStatsNegativeCount(t *testing.T) {
	stats := Stats{
		Count:       -5,
		TotalTimeUs: 23998,
	}
	_, err := stats.JsonSummary()
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
}

const testPort int = 8080

func getTestUrl(endpoint string) string {
	return fmt.Sprintf("http://localhost:%d%s", testPort, endpoint)
}

func TestServerStartAndShutdown(t *testing.T) {
	server, err := StartServer(testPort, 0)
	if err != nil {
		t.Error(err)
	}

	server.Shutdown()
	<-server.Done
}

func TestStartServerPortAlreadyBound(t *testing.T) {
	server, err := StartServer(testPort, 0)
	if err != nil {
		t.Error(err)
	}
	duplicateServer, err := StartServer(testPort, 0)
	if err == nil {
		t.Error("Expected error to be non-nil")
	}
	if duplicateServer != nil {
		t.Error("Expected duplicateServer to be nil")
	}

	server.Shutdown()
	<-server.Done
}

func TestStartServerNegativePort(t *testing.T) {
	server, err := StartServer(-26, 0)
	if err == nil {
		t.Error("Expected error to be non-nil")
	}
	if server != nil {
		t.Error("Expected server to be nil")
	}
}

func TestServerShutdownTwice(t *testing.T) {
	server, err := StartServer(testPort, 0)
	if err != nil {
		t.Error(err)
	}

	server.Shutdown()
	server.Shutdown()
	<-server.Done
}

func TestHashPost(t *testing.T) {
	server, err := StartServer(testPort, time.Nanosecond)
	if err != nil {
		t.Error(err)
	}
	client := &http.Client{}

	postResponse, err := client.PostForm(
		getTestUrl("/hash"),
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
	server, err := StartServer(8080, time.Nanosecond)
	if err != nil {
		t.Error(err)
	}
	client := &http.Client{}

	postResponse, err := client.PostForm(
		getTestUrl("/hash"), url.Values{})
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
	server, err := StartServer(8080, time.Nanosecond)
	if err != nil {
		t.Error(err)
	}
	client := &http.Client{}

	getResponse, err := client.Get(getTestUrl("/hash"))
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

	pw := "PeachPie"
	postResponse, err := client.PostForm(
		getTestUrl("/hash"),
		url.Values{"password": {pw}})
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

	getResponse, err := client.Get(getTestUrl(fmt.Sprintf("/hash/%s", id)))
	if err != nil {
		t.Error(err)
	}
	hash, err := ioutil.ReadAll(getResponse.Body)
	getResponse.Body.Close()
	if err != nil {
		t.Error(err)
	}
	expectedHash, err := HashPassword(pw)
	if err != nil {
		t.Error(err)
	}
	if string(hash) != expectedHash.String() {
		t.Errorf("For password %s got hash %s, expected %s\n",
			pw, hash, expectedHash)
	}

	shutdownResponse, err := client.Get("http://localhost:8080/shutdown")
	if err != nil {
		t.Error(err)
	}
	io.Copy(ioutil.Discard, shutdownResponse.Body)
	shutdownResponse.Body.Close()

	<-server.Done
}

func BenchmarkPostHash(b *testing.B) {
	server, err := StartServer(8080, 5000)
	if err != nil {
		b.Error(err)
	}
	client := &http.Client{}
	formValues := url.Values{"password": {"RogerRoger"}}
	for i := 0; i < b.N; i++ {
		resp, err := client.PostForm("http://localhost:8080/hash", formValues)
		if err != nil {
			b.Error(err)
		}
		io.Copy(ioutil.Discard, resp.Body)
		resp.Body.Close()
	}

	server.Shutdown()
	<-server.Done
}
