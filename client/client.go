package missy

import (
	"net/http"
	"encoding/json"
	"github.com/pkg/errors"
	"bytes"
	"strings"
	"fmt"
	"time"
	"github.com/microdevs/missy/log"
	gourl "net/url"
)

const ContentTypeJson = "application/json"

type Client interface {
	Create(entity string, data interface{}) error
	Update(entity string, id string, data interface{}) error
	Delete(entity string, id string) error
	Get(entity string, id string, v interface{}) error
	GetList(entity string, limit int, skip int, filters map[string]string) (interface{}, error)
}

type Connection interface {
	Call(method string, path string, query gourl.Values, v interface{}) (*http.Response, error)
}

// todo: implement token auth
// MiSSy Service Connection object. Target is the service that will be called
type ServiceConnection struct {
	Target string
	Host string
	Port string
	Token string
	Timeout int
}

// Create a new ServiceConnection for a target service
func NewConnection(target string) *ServiceConnection {
	//todo: implement service discovery and set host and port
	return &ServiceConnection{Target: target}
}

// Generall Call function can handle all http methods, query string and json marshalling for POST and PUT
func (c *ServiceConnection) Call(method string, path string, query gourl.Values, v interface{}) (*http.Response, error) {
	//  create a reader in case we need to post or put
	var dataReader *bytes.Reader

	if method == http.MethodPost || method == http.MethodPut {
		dataByte, jsonErr := json.Marshal(v)
		if jsonErr != nil {
			return nil, errors.New("Failed to Marshal data to json")
		}
		dataReader = bytes.NewReader(dataByte)
	}

	// build a querystring if there are some values
	querystring := ""
	if len(query) > 0 {
		querystring = "?"+query.Encode()
	}

	// build the url
	url := "http://"+c.Target+":8080"+"/"+strings.TrimLeft(path, "/") + querystring
	log.Debugf("Calling service %s with url %s and method %s", c.Target, url, method)
	// create a net/http client
	// todo: make timeout configurable
	client := http.Client{Timeout: 10 * time.Second}
	// make the differnt calls depending on the method
	switch method {
		case http.MethodGet:
			return client.Get(url)
		case http.MethodPost:
			return client.Post(url, ContentTypeJson, dataReader)
		case http.MethodHead:
			return client.Head(url)
		case http.MethodPut:
			request, err := http.NewRequest(http.MethodPut, url, dataReader)
			if err != nil {
				return nil, errors.New(fmt.Sprintf("unable to build request with error %s", err))
			}
			request.Header.Set("Content-Type", ContentTypeJson)
			return client.Do(request)
		default:
			return nil, errors.New(fmt.Sprintf("Method %s not implemented", method))
	}
}



