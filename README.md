# hasher
A password hashing web service

## Getting Started
### Build

To build the main hasher executable simply clone the repo and run go build

    $ git clone https://github.com/jjkr/hasher.git
    $ cd hasher
    $ go build

### Run

The resulting hasher executable can be run with a single argument, the port to bind to

    $./hasher 8080

## API

   POST /hash
   GET  /hash/:id
   GET  /stats
   POST /shutdown
