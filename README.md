# MIcroservice Support SYstem (MISSY)

[![Build Status](https://travis-ci.org/microdevs/missy.svg?branch=master)](https://travis-ci.org/microdevs/missy) [![Coverage Status](https://coveralls.io/repos/github/microdevs/missy/badge.svg?branch=master)](https://coveralls.io/github/microdevs/missy?branch=master)

MISSY is a library for creating REST services that talk to each other. It provides the following functionality

### Features

* Routing with gorrila/mux
* Logging
* Configuration with environment variables
* Monitoring with Prometheus
* /info /health 

### Roadmap

* Service discovery
* REST client
* Security

## How to use it

Example for a simple hello world service

```
# main.go

package main

import (
	"github.com/microdevs/missy/server"
	"net/http"
	"fmt"
)

func main() {
	s := server.NewServer("hello-service")
	s.HandleFunc("/foo/{name}", fooHandler).Methods("GET")
	s.Port = "8088"
	s.StartServer()
}

func fooHandler(w http.ResponseWriter, r *http.Request) {

	vars := server.Vars(r)
	w.Write([]byte(fmt.Sprintf("Hello %s", vars["name"])))
}

```

### Run it:
```go run main.go```

### Call the Endpoint:
```
curl "http://localhost:8088/foo/bar"
```

Get Prometheus Metrics:
```
curl http://localhost:8088/metrics
```

### Get Info:
```
http://localhost:8088/info
```

Response:
```
Name service-a
Uptime 14.504883092s
```
### Get Health:
```
http://localhost:8088/health
```

Response:
```
OK
```
