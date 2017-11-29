package client

import (
	"net/http"
	"encoding/json"
	"github.com/pkg/errors"
	"bytes"
	"strings"
	"fmt"
	"time"
	"io/ioutil"
	"strconv"
	"github.com/microdevs/missy/log"
	gourl "net/url"
)

const ContentTypeJson = "application/json"

// todo: implement token auth
// MiSSy Client object. Target is the service that will be called
type Client struct {
	Target string
	host string
	port string
}

// Create a new Client for a target service
func New(target string) *Client {
	//todo: implement service discovery and set host and port
	return &Client{Target: target}
}



// Generall Call function can handle all http methods, query string and json marshalling for posts
func (c *Client) Call(method string, path string, query gourl.Values, v interface{}) (*http.Response, error) {
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
	log.Debugf("Calling service %s with url %s", c.Target, url)
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

// create wrapper to post a new entry
func (c *Client) Create(entity string, data interface{}) error {
	resp, callErr := c.Call(http.MethodPost, entity, nil, data)
	if callErr != nil {
		return errors.New(fmt.Sprintf("Error calling service %s with message %s", c.Target, callErr))
	}
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		return errors.New(fmt.Sprintf("Error creating entity with code %s and message %s", string(resp.StatusCode), string(bodyBytes)))
	}
	return nil
}

// update wrapper to put a entry
func (c *Client) Update(entity string, id string, data interface{}) error {
	resp, callErr := c.Call(http.MethodPost, entity+"/"+id, nil, data)
	if callErr != nil {
		return errors.New(fmt.Sprintf("Error calling service %s with message %s", c.Target, callErr))
	}
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		return errors.New(fmt.Sprintf("Error updating entity with code %s and message %s", string(resp.StatusCode), string(bodyBytes)))
	}
	return nil
}

// get entity by id from service
func (c *Client) GetById(entity string, id string, v interface{}) error {
	resp, callErr := c.Call(http.MethodGet, entity+"/"+id, nil, nil)
	if callErr != nil {
		return errors.New(fmt.Sprintf("Error calling service %s with message %s", c.Target, callErr))
	}
	bodyBytes, readErr := ioutil.ReadAll(resp.Body)
	if readErr != nil {
		return errors.New(fmt.Sprintf("Error Reading Body from service %s with message %s", c.Target, readErr))
	}
	json.Unmarshal(bodyBytes, v)
	return nil
}

// get entity list from service
func (c *Client) GetList(entity string, limit int, skip int, v interface{}) error {
	var query gourl.Values  = gourl.Values{}
	query.Set("limit", strconv.Itoa(limit))
	query.Set("skip", strconv.Itoa(skip))
	querystring := query.Encode()
	log.Debugf("Query %s", querystring)
	resp, callErr := c.Call(http.MethodGet, entity+"/list", query, nil)
	if callErr != nil {
		return errors.New(fmt.Sprintf("Error calling service %s with message %s", c.Target, callErr))
	}
	bodyBytes, readErr := ioutil.ReadAll(resp.Body)
	if readErr != nil {
		return errors.New(fmt.Sprintf("Error Reading Body from service %s with message %s", c.Target, readErr))
	}
	json.Unmarshal(bodyBytes, v)
	return nil
}

