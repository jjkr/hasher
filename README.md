# Hasher
A password hashing web service

## Getting Started

### Build

To build the main hasher executable simply clone the repo and run ```go build```.

    $ git clone https://github.com/jjkr/hasher.git
    $ cd hasher
    $ go build

### Test

Unit tests are run with ```go test```

    $ go test
    PASS
    ok      github.com/jjkr/hasher  0.025s

### Run

The resulting hasher executable can be run with a single argument, the port to bind to

    $ ./hasher 8080

## API

### POST /hash

Creates a new hash and returns the id of the newly created hash. The hash is not available immediately. Instead, clients have to poll the GET endpoint until the hash is available. Hash id increases with each request.

Returns: 202 Accepted with the hash's id in the body of the response. A hash id is a 128 bit hex encoded value

Example:

    $ curl -v --data "password=IceCream" http://localhost:5000/hash
    *   Trying ::1...
    * TCP_NODELAY set
    * Connected to localhost (::1) port 5000 (#0)
    > POST /hash HTTP/1.1
    > Host: localhost:5000
    > User-Agent: curl/7.54.0
    > Accept: */*
    > Content-Length: 17
    > Content-Type: application/x-www-form-urlencoded
    >
    * upload completely sent off: 17 out of 17 bytes
    < HTTP/1.1 202 Accepted
    < Date: Sun, 20 May 2018 23:07:17 GMT
    < Content-Length: 32
    < Content-Type: text/plain; charset=utf-8
    <
    * Connection #0 to host localhost left intact
    15307cee701bf5617fc3fedb06e324de

### GET /hash/:id

Get the hash for the given id. Returns a base64 encoded hash value or 404 if not found or if the hash has not finished computing.

Example:

    $ curl -v http://localhost:5000/hash/15307cee701bf5617fc3fedb06e324de
    *   Trying ::1...
    * TCP_NODELAY set
    * Connected to localhost (::1) port 5000 (#0)
    > GET /hash/15307cee701bf5617fc3fedb06e324de HTTP/1.1
    > Host: localhost:5000
    > User-Agent: curl/7.54.0
    > Accept: */*
    > 
    < HTTP/1.1 200 OK
    < Date: Sun, 20 May 2018 23:14:57 GMT
    < Content-Length: 88
    < Content-Type: text/plain; charset=utf-8
    < 
    * Connection #0 to host localhost left intact
    iwUTRA9dzx5DPVVTDk12e1ZzgGvKf3Bl56YJDruE92sPwk9VYEJ4gT9h+FK/941r17Ecq67YLjvjnikdpnfpZA==

### GET /stats

Get statistics about hash requests. Returns a JSON object with the total number of valid POST hash requests and the average amount of time it took to process each request in microseconds. The processing time does not include the actual password hashing, just the time to handle the POST request.

Example:

    $ curl -v http://localhost:8080/stats
    *   Trying ::1...
    * TCP_NODELAY set
    * Connected to localhost (::1) port 8080 (#0)
    > GET /stats HTTP/1.1
    > Host: localhost:8080
    > User-Agent: curl/7.54.0
    > Accept: */*
    >
    < HTTP/1.1 200 OK
    < Date: Mon, 21 May 2018 01:06:37 GMT
    < Content-Length: 45
    < Content-Type: text/plain; charset=utf-8
    <
    * Connection #0 to host localhost left intact
    {"total":700000,"average":23.403601428571427}

### POST /shutdown

Shutdown the webserver, waiting for all active requests to complete.

Returns 200 OK

Example:

    $ curl -v http://localhost:8080/shutdown
    *   Trying ::1...
    * TCP_NODELAY set
    * Connected to localhost (::1) port 8080 (#0)
    > GET /shutdown HTTP/1.1
    > Host: localhost:8080
    > User-Agent: curl/7.54.0
    > Accept: */*
    >
    < HTTP/1.1 200 OK
    < Date: Tue, 22 May 2018 00:47:44 GMT
    < Content-Length: 0
    <
    * Connection #0 to host localhost left intact


Copyright Joe Kramer 2018. All rights reserved
