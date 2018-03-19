package data

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

type TestStruct struct {
	A int
	B string
	C []string
}

var subject = TestStruct{
	1,
	"foo",
	[]string{"foo", "bar", "baz"},
}

var expectedJSON = `{"A":1,"B":"foo","C":["foo","bar","baz"]}`
var expectedXML = `<TestStruct><A>1</A><B>foo</B><C>foo</C><C>bar</C><C>baz</C></TestStruct>`

func TestMarshalToJson(t *testing.T) {

	request := &http.Request{
		Header: http.Header{
			httpHeaderAccept: []string{contentTypeJSON},
		},
	}

	rec := httptest.NewRecorder()

	Marshal(rec, request, subject)

	if rec.Code != http.StatusOK {
		t.Errorf("Marshaller failed with an unknown error")
	}

	response := rec.Result()

	if ct := response.Header.Get(httpHeaderContentType); ct != contentTypeJSON {
		t.Errorf("Accept header has the wrong content type. Expected %s, actual %s", contentTypeJSON, ct)
	}
	bodyBytes, _ := ioutil.ReadAll(rec.Body)
	if strBody := string(bodyBytes); strBody != expectedJSON {
		t.Errorf("Body does not match expected output. Expected %s, actual %s", expectedJSON, strBody)
	}

}

func TestMarshalToXML(t *testing.T) {

	request := &http.Request{
		Header: http.Header{
			httpHeaderAccept: []string{contentTypeApplicationXML},
		},
	}

	rec := httptest.NewRecorder()

	Marshal(rec, request, subject)

	response := rec.Result()

	if rec.Code != http.StatusOK {
		t.Error("Marshaller failed with an unknown error")
	}

	if ct := response.Header.Get(httpHeaderContentType); ct != contentTypeApplicationXML {
		t.Errorf("Accept header has the wrong content type. Expected %s, actual %s", contentTypeApplicationXML, ct)
	}
	bodyBytes, _ := ioutil.ReadAll(rec.Body)
	if strBody := string(bodyBytes); strBody != expectedXML {
		t.Errorf("Body does not match expected output. Expected %s, actual %s", expectedXML, strBody)
	}

}
