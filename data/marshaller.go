package data

import (
	"net/http"
	"reflect"
	"encoding/xml"
	"encoding/json"
	"github.com/microdevs/missy/log"
)

func MarshalResponse(w http.ResponseWriter, r *http.Request, subject interface{}) ([]byte, error) {

	var resp []byte
	var err error
	var convertTo string

	if r.Header.Get("Accept") == "text/xml" {
		convertTo = "xml"
		// todo: if it's a pointer follow the pointer and use the data
		if reflect.TypeOf(subject).Kind() == reflect.Slice {
			s := reflect.ValueOf(subject)
			interfaceSlice := make([]interface{}, s.Len())
			for i:=0; i<s.Len(); i++ {
				interfaceSlice[i] = s.Index(i).Interface()
			}
			wrapper := Results{}
			wrapper.Results = interfaceSlice
			wrapper.Length = len(interfaceSlice)
			resp, err = xml.Marshal(wrapper)
		} else {
			resp, err = xml.Marshal(subject)
		}

		w.Header().Set("Content-Type", "text/xml")
	} else {
		convertTo = "json"
		resp, err = json.Marshal(subject)
		w.Header().Set("Content-Type", "application/json")
	}

	if err != nil {
		log.Errorf("Error marshalling creditor to %s: %v", convertTo, err)
	}

	return resp, err
}

type Results struct {
	XMLName xml.Name `xml:"result"`
	Results []interface{}
	Length int `xml:"length,attr"`
}