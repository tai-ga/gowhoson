### gowhoson

gowhoson is a golang implementation of the "Whoson" protocol.

[![Build Status](https://travis-ci.org/tai-ga/gowhoson.svg?branch=master)](https://travis-ci.org/tai-ga/gowhoson)
[![codecov](https://codecov.io/gh/tai-ga/gowhoson/branch/master/graph/badge.svg)](https://codecov.io/gh/tai-ga/gowhoson)
[![Go Report Card](https://goreportcard.com/badge/github.com/tai-ga/gowhoson)](https://goreportcard.com/report/github.com/tai-ga/gowhoson)
[![GoDoc](https://godoc.org/github.com/tai-ga/gowhoson/whoson?status.svg)](http://godoc.org/github.com/tai-ga/gowhoson/whoson)
[![GitHub release](https://img.shields.io/github/release/tai-ga/gowhoson.svg)](https://github.com/tai-ga/gowhoson/releases/latest)
[![MIT License](http://img.shields.io/badge/license-MIT-blue.svg?style=flat)](https://github.com/tai-ga/gowhoson/blob/master/LICENSE)

#### What is whoson ?
Whoson ("WHO iS ONline") is a proposed Internet protocol that allows Internet server programs know if a particular (dynamically allocated) IP address is currently allocated to a known (trusted) user and, optionally, the identity of the said user.
The protocol could be used by an SMTP Message Transfer System in conjunction with anti-spam-relaying filters to implement a scheme similar to the one described here to allow roaming customers use their "home" SMTP server to submit email while connected from a "foreign" network.

#### Link
* "Whoson" Project page. [http://whoson.sourceforge.net/](http://whoson.sourceforge.net/).
* About the "Whoson" protocol.  [http://whoson.sourceforge.net/whoson.txt](http://whoson.sourceforge.net/whoson.txt)

#### Examples whoson package
Server 01
```go
func main() {
        whoson.ListenAndServe("tcp", ":9876")
}
```
Server 02
```go
func main() {
  addr := net.TCPAddr{
    Port: 9876,
    IP:   net.ParseIP("127.0.0.1"),
  }

  l, err := net.ListenTCP("tcp", &addr)
  if err != nil {
    log.Fatalf("%v", err)
  }
  whoson.ServeTCP(l)
}
```

Client
```go
func main() {
        client, err := whoson.Dial("tcp", "127.0.0.1:9876")
        if err != nil {
                log.Fatalf("%v", err)
        }
        defer client.Quit()

        res, err := client.Login("192.168.0.1", "user01@example.org")
        if err != nil {
                log.Fatalf("%v", err)
        }
        fmt.Println(res.String())
}
```
#### Install
```
$ go get -u github.com/tai-ga/gowhoson/cmd/gowhoson
```

#### Usage
Server
```
> gowhoson server -h
NAME:
   gowhoson server - gowhoson server mode

USAGE:
   gowhoson server [command options] [arguments...]

OPTIONS:
   --tcp value       e.g. [ServerIP:Port|nostart] [$GOWHOSON_SERVER_TCP]
   --udp value       e.g. [ServerIP:Port|nostart] [$GOWHOSON_SERVER_UDP]
   --log value       e.g. [stdout|stderr|discard] or "/var/log/filename.log" [$GOWHOSON_SERVER_LOG]
   --loglevel value  e.g. [debug|info|warn|error|dpanic|panic|fatal] [$GOWHOSON_SERVER_LOGLEVEL]
   --serverid value  e.g. [1000] (default: 0) [$GOWHOSON_SERVER_SERVERID]
   --expvar          e.g. (default: false) [$GOWHOSON_SERVER_EXPVAR]
```

Client
```
> gowhoson client -h
NAME:
   gowhoson client - gowhoson client mode

USAGE:
   gowhoson client command [command options] [arguments...]

COMMANDS:
     login       whoson command "LOGIN"
     query       whoson command "QUERY"
     logout      whoson command "LOGOUT"
     editconfig  edit client configration file

OPTIONS:
   --help, -h  show help
```
#### Implemented commands

* LOGIN
* LOGOUT
* QUERY
* QUIT

#### Reference

* Original reference implementation of whoson.
* Many japanese gopher products.
* :tada: Many Thanks! :tada:

#### Contribute

1. fork a repository: github.com/tai-ga/gowhoson to github.com/you/repo
2. get original code: go get github.com/tai-ga/gowhoson
3. work on original code
4. add remote to your repo: git remote add myfork https://github.com/you/repo.git
5. push your changes: git push myfork
6. create a new Pull Request

- see [GitHub and Go: forking, pull requests, and go-getting](http://blog.campoy.cat/2014/03/github-and-go-forking-pull-requests-and.html)

#### License

MIT

#### Author

Masahiro Ono ([@tai-ga](https://twitter.com/tai_ga))
