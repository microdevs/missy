package data

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"reflect"

	"github.com/microdevs/missy/log"
)

const httpHeaderAccept = "Accept"
const httpHeaderContentType = "Content-Type"
const contentTypeJSON = "application/json"
const contentTypeTexXML = "text/xml"
const contentTypeApplicationXML = "application/xml"

// Marshal will marshal any interface{} according to the Accept header of the passed request to JSON by default or XML if the header is set to text/xml
func Marshal(w http.ResponseWriter, r *http.Request, data interface{}) {
	MarshalWithCode(w, r, data, http.StatusOK)
}

// MarshalWithCode will marshal any interface{} according to the Accept header and will use the statusCode provided
func MarshalWithCode(w http.ResponseWriter, r *http.Request, data interface{}, statusCode int) {

	var resp []byte
	var err error
	var contentType string

	switch r.Header.Get(httpHeaderAccept) {
	case contentTypeTexXML, contentTypeApplicationXML:
		contentType = contentTypeApplicationXML
		// todo: if it's a pointer follow the pointer and use the data
		if reflect.TypeOf(data).Kind() == reflect.Slice {
			s := reflect.ValueOf(data)
			interfaceSlice := make([]interface{}, s.Len())
			for i := 0; i < s.Len(); i++ {
				interfaceSlice[i] = s.Index(i).Interface()
			}
			wrapper := Results{}
			wrapper.Results = interfaceSlice
			wrapper.Length = len(interfaceSlice)
			resp, err = xml.Marshal(wrapper)
		} else {
			resp, err = xml.Marshal(data)
		}
	default:
		contentType = contentTypeJSON
		resp, err = json.Marshal(data)
	}

	if err != nil {
		log.Errorf("Error marshalling to %s: %v", contentType, err)
		http.Error(w, fmt.Sprintf("Error marshalling object to %s: %s", contentType, err), http.StatusInternalServerError)
	}

	w.Header().Set(httpHeaderContentType, contentType)
	w.WriteHeader(statusCode)
	w.Write(resp)
}

// Results is a wrapper type to wrap results in an XML <result> node
type Results struct {
	XMLName xml.Name `xml:"result"`
	Results []interface{}
	Length  int `xml:"length,attr"`
}
