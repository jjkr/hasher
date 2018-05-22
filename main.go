package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"
)

func showUsage() {
	fmt.Printf("usage: %s PORT\n", os.Args[0])
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("ERROR too few arguments")
		showUsage()
		os.Exit(-1)
	}

	rand.Seed(time.Now().UTC().UnixNano())

	port, err := strconv.Atoi(os.Args[1])
	if err != nil {
		fmt.Printf("ERROR bad port: %s\n", os.Args[1])
		showUsage()
		os.Exit(-1)
	}
	server, err := StartServer(port, 5*time.Second)
	if err != nil {
		log.Printf("Error starting server: %s", err)
		os.Exit(-1)
	}
	log.Printf("Starting hasher server on port %d", port)
	<-server.Done
}
