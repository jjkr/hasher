# hasher
A password hashing web service

## Getting Started
### Build

To build the main hasher executable simply clone the repo and run go build.

    $ git clone https://github.com/jjkr/hasher.git
    $ cd hasher
    $ go build

### Run

The resulting hasher executable can be run with a single argument, the port to bind to

    $./hasher 8080

The server will listen on all available interfaces.

## API

###POST /hash

Creates a new hash and returns the id of the newly created hash. The hash is not available immediately. Instead, clients have to poll the get endpoint until the id is available.

Returns: 202 Accepted and the hash's id in the body of the response. A hash id is a 128 bit hex encoded value

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

###GET  /hash/:id

Get the hash for the given id. Returns a base64 encoded hash value or 404 if not found.

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

###GET  /stats

Get statistics about hash requests. Returns a JSON object with the total number of successful requests to create a hash and the average amount of time it took to process each request in microseconds.


###POST /shutdown
