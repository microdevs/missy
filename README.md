# MIcroservice Support SYstem (MiSSy)

[![Slack Status](https://microdevs-slackin.herokuapp.com/badge.svg)](https://microdevs-slackin.herokuapp.com) [![Build Status](https://travis-ci.org/microdevs/missy.svg?branch=master)](https://travis-ci.org/microdevs/missy) [![Coverage Status](https://coveralls.io/repos/github/microdevs/missy/badge.svg?branch=master)](https://coveralls.io/github/microdevs/missy?branch=master) [![Go Report Card](https://goreportcard.com/badge/github.com/microdevs/missy)](https://goreportcard.com/report/github.com/microdevs/missy)


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

Create a `.missy.yml` config file in the root directory of your service with the following content

```
name: hello
```

```
# main.go

package main

import (
	"github.com/microdevs/missy/service"
	"net/http"
	"fmt"
)

func main() {
	s := service.New()
	s.HandleFunc("/hello/{name}", HelloHandler).Methods("GET")
	s.Start()
}

func HelloHandler(w *service.ResponseWriter, r *http.Request) {
	vars := server.Vars(r)
	w.Write([]byte(fmt.Sprintf("Hello %s", vars["name"])))
}

```

### Run it:
```go run main.go```

### Call the Endpoint:
```
curl "http://localhost:8088/hello/microdevs"
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
Name hello
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
