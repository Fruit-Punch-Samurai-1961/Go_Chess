## Go Chess
Go Chess is a full-stack app that uses WebSockets to allow users to play chess with each other.

## Prerequisites
Before you begin, ensure that you have the following requirements:

* [Golang](https://golang.org/dl/)
* [mySQL](https://dev.mysql.com/download)
* Following Javascript libraries:
    * [chess.js](https://github.com/jhlywa/chess.js/)
    * [chessboardjs](https://chessboardjs.com/download)
* An HTTPS Certificate:
    * If this is only for testing, you can generate your own using: <br />
      `go run %GOROOT%/src/crypto/tls/generate_cert.go --host=localhost`
    * Note: The default assumes that you have a `tls` directory under which `cert.pem` and `key.pem` are located
## Installation
To install Go Chess, run the following command:
`go get github.com/sheshan1961/chessapp`


## Usage
1. Make a database called `chessapp` and its respective tables using the files located under `chessapp/mySQL_code`
2. Run `go build web/`
    * You can set three flags: <br />
        * `addr`-HTTP network address
        * `static-dir`-Path to Static Dir
        * `dsn`-MySQL Database for chess games(Set username and password)
