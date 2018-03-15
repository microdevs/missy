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
const contentTypeXML = "text/xml"

// Marshal will marshal any interface{} according to the Accept header of the passed request to JSON by default or XML if the header is set to text/xml
func Marshal(w http.ResponseWriter, r *http.Request, subject interface{}) {

	var resp []byte
	var err error
	var convertTo string

	if r.Header.Get(httpHeaderAccept) == contentTypeXML {
		convertTo = "xml"
		// todo: if it's a pointer follow the pointer and use the data
		if reflect.TypeOf(subject).Kind() == reflect.Slice {
			s := reflect.ValueOf(subject)
			interfaceSlice := make([]interface{}, s.Len())
			for i := 0; i < s.Len(); i++ {
				interfaceSlice[i] = s.Index(i).Interface()
			}
			wrapper := Results{}
			wrapper.Results = interfaceSlice
			wrapper.Length = len(interfaceSlice)
			resp, err = xml.Marshal(wrapper)
		} else {
			resp, err = xml.Marshal(subject)
		}

		w.Header().Set(httpHeaderContentType, contentTypeXML)
		w.Write(resp)
		return
	}
	convertTo = "json"
	resp, err = json.Marshal(subject)
	if err != nil {
		log.Errorf("Error marshalling to %s: %v", convertTo, err)
		http.Error(w, fmt.Sprintf("Error marshalling object to %s: %s", convertTo, err), http.StatusInternalServerError)
	}

	w.Header().Set(httpHeaderContentType, contentTypeJSON)
	w.Write(resp)
}

// Results is a wrapper type to wrap results in an XML <result> node
type Results struct {
	XMLName xml.Name `xml:"result"`
	Results []interface{}
	Length  int `xml:"length,attr"`
}
