# MiSSy

[![Slack Status](https://microdevs-slackin.herokuapp.com/badge.svg)](https://microdevs-slackin.herokuapp.com) [![Build Status](https://travis-ci.org/microdevs/missy.svg?branch=master)](https://travis-ci.org/microdevs/missy) [![Coverage Status](https://coveralls.io/repos/github/microdevs/missy/badge.svg?branch=master)](https://coveralls.io/github/microdevs/missy?branch=master) [![Go Report Card](https://goreportcard.com/badge/github.com/microdevs/missy)](https://goreportcard.com/report/github.com/microdevs/missy)

MiSSy (Microservice Support System) is an SDK for creating REST services that talk to each other. It provides the following functionality:

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

```go
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
	s.HandleFunc("/json", JsonHandler).Methods("GET")
	s.Start()
}

func HelloHandler(w http.ResponseWriter, r *http.Request) {
	vars := service.Vars(r)
	w.Write([]byte(fmt.Sprintf("Hello %s", vars["name"])))
}

// marshalling example
type MyType struct {
	A int
	B string
}

func JsonHandler(w http.ResponseWriter, r *http.Request) {
	mytype := MyType{
		A: 123,
		B: "Hello world",
	}
	
	data.Marshal(w, r, mytype)
}

// response body {A:123,B:"Hello world"}

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

### Messaging
Use messaging.Reader and messaging.Writer to subscribe and publish messages.
It uses kafka underneath.
Example usage:

Reader with brokers hosts, group-id and topic
```
reader := messaging.NewReader([]string{"localhost:9092"}, "group-id", "topic")
err := reader.Read(func(msg Message) error{
    // do something with msg
    
    // return nil or error (if commit should not happen)
})

// remember to close reader after use
defer reader.Close()
```

Writer with brokers hosts and topic
```
writer := messaging.NewWriter([]string{"localhost:9092"}, "topic")
err := writer.Write([]byte("key"), []byte("value"))

// remember to close writer after use
defer writer.Close()
```
