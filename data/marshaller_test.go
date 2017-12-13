package data

import (
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

	resp, err := MarshalResponse(rec, request, subject)

	if err != nil {
		t.Errorf("Failed to call marshalling function with error %s", err)
	}

	response := rec.Result()

	if ct := response.Header.Get(httpHeaderContentType); ct != contentTypeJSON {
		t.Errorf("Accept header has the wrong content type. Expected %s, actual %s", contentTypeJSON, ct)
	}

	if strbody := string(resp); strbody != expectedJSON {
		t.Errorf("Body does not match expected output. Expected %s, actual %s", expectedJSON, strbody)
	}

}

func TestMarshalToXML(t *testing.T) {

	request := &http.Request{
		Header: http.Header{
			httpHeaderAccept: []string{contentTypeXML},
		},
	}

	rec := httptest.NewRecorder()

	resp, err := MarshalResponse(rec, request, subject)

	if err != nil {
		t.Errorf("Failed to call marshalling function with error %s", err)
	}

	response := rec.Result()

	if ct := response.Header.Get(httpHeaderContentType); ct != contentTypeXML {
		t.Errorf("Accept header has the wrong content type. Expected %s, actual %s", contentTypeXML, ct)
	}

	if strbody := string(resp); strbody != expectedXML {
		t.Errorf("Body does not match expected output. Expected %s, actual %s", expectedXML, strbody)
	}

}
