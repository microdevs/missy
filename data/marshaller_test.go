package data

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

type TestCases []TestCase

type TestCase struct {
	Data                interface{}
	Accept              string
	ExpectedContentType string
	ExpectedOutput      string
}

type Data struct {
	A int
	B string
	C []string
}

var testCases = TestCases{
	{
		Data: Data{
			1,
			"foo",
			[]string{"foo", "bar", "baz"},
		},
		Accept:              contentTypeApplicationXML,
		ExpectedContentType: contentTypeApplicationXML,
		ExpectedOutput:      `<Data><A>1</A><B>foo</B><C>foo</C><C>bar</C><C>baz</C></Data>`,
	},
	{
		Data: Data{
			1,
			"foo",
			[]string{"foo", "bar", "baz"},
		},
		Accept:              contentTypeTextXML,
		ExpectedContentType: contentTypeApplicationXML,
		ExpectedOutput:      `<Data><A>1</A><B>foo</B><C>foo</C><C>bar</C><C>baz</C></Data>`,
	},
	{
		Data: Data{
			1,
			"foo",
			[]string{"foo", "bar", "baz"},
		},
		Accept:              contentTypeJSON,
		ExpectedContentType: contentTypeJSON,
		ExpectedOutput:      `{"A":1,"B":"foo","C":["foo","bar","baz"]}`,
	},
	{
		Data:                []string{"foo", "bar", "baz"},
		Accept:              contentTypeApplicationXML,
		ExpectedContentType: contentTypeApplicationXML,
		ExpectedOutput:      `<result length="3"><Results>foo</Results><Results>bar</Results><Results>baz</Results></result>`,
	},
}

func TestMarshal(t *testing.T) {

	for _, tc := range testCases {
		request := httptest.NewRequest("GET", "http://missy.local/marshal", nil)
		request.Header.Set("Accept", tc.Accept)
		response := httptest.NewRecorder()
		Marshal(response, request, tc.Data)
		if ct := response.Header().Get(httpHeaderContentType); ct != tc.ExpectedContentType {
			t.Logf("Accept header has the wrong content type. Expected %s, actual %s", tc.ExpectedContentType, ct)
			t.Fail()
		}
		if response.Code != http.StatusOK {
			t.Log("Marshaller returned not StatusCode 200 OK")
			t.Fail()
		}
		bodyBytes, _ := ioutil.ReadAll(response.Body)
		if strBody := string(bodyBytes); strBody != tc.ExpectedOutput {
			t.Errorf("Body does not match expected output. Expected %s, actual %s", tc.ExpectedOutput, strBody)
		}
	}
}

func TestMarshalWithCode(t *testing.T) {
	for _, tc := range testCases {
		request := httptest.NewRequest("GET", "http://missy.local/marshal", nil)
		request.Header.Set("Accept", tc.Accept)
		response := httptest.NewRecorder()
		MarshalWithCode(response, request, tc.Data, http.StatusTeapot)
		if ct := response.Header().Get(httpHeaderContentType); ct != tc.ExpectedContentType {
			t.Logf("Accept header has the wrong content type. Expected %s, actual %s", tc.ExpectedContentType, ct)
			t.Fail()
		}
		if response.Code != http.StatusTeapot {
			t.Log("Marshaller returned not StatusCode 418 I am a Teapot")
			t.Fail()
		}
		bodyBytes, _ := ioutil.ReadAll(response.Body)
		if strBody := string(bodyBytes); strBody != tc.ExpectedOutput {
			t.Errorf("Body does not match expected output. Expected %s, actual %s", tc.ExpectedOutput, strBody)
		}
	}
}
