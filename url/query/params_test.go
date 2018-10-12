package query_test

import (
	"net/url"
	"reflect"
	"testing"

	. "github.com/microdevs/missy/url/query"
)

var testParams = []struct {
	in  string
	out Values
}{
	{"http://localhost?value=gt:10", Values{"value": Operators{GT: []string{"10"}}}},
	{"http://localhost?value=gt:10&value=lte:100", Values{"value": Operators{GT: []string{"10"}, LTE: []string{"100"}}}},
	{"http://localhost?sort=asc:key1", Values{"sort": Operators{Asc: []string{"key1"}}}},
	{"http://localhost?sort=desc:key1,key2,key3", Values{"sort": Operators{Desc: []string{"key1"}, Default: []string{"key2", "key3"}}}},
	{"http://localhost?sort=desc:key1,key2,asc:key3", Values{"sort": Operators{Asc: []string{"key3"}, Desc: []string{"key1"}, Default: []string{"key2"}}}},
	{"http://localhost?date=gt:2018-01-01&date=lte:2018-04-01", Values{"date": Operators{GT: []string{"2018-01-01"}, LTE: []string{"2018-04-01"}}}},
	{"http://localhost?name=name&first_value=100&desc_value=inner", Values{"name": Operators{Default: []string{"name"}}, "first_value": Operators{Default: []string{"100"}}, "desc_value": Operators{Default: []string{"inner"}}}},
}

func TestParse(t *testing.T) {
	for _, tt := range testParams {
		t.Run(tt.in, func(t *testing.T) {

			parsedURL, err := url.Parse(tt.in)
			if err != nil {
				t.Errorf("cannot parse raw url: %v", err)
			}

			queryParams, err := url.ParseQuery(parsedURL.RawQuery)
			if err != nil {
				t.Errorf("cannot parse query: %v", err)
			}

			queryValues := Parse(queryParams)
			if len(tt.out) != len(queryValues) {
				t.Errorf("invalid query paramterers Parsing len, expected: %v = %v", len(tt.out), len(queryValues))
			}

			for paramKey, values := range queryValues {
				if _, ok := tt.out[paramKey]; !ok {
					t.Errorf("Values paramKey key not exist: %s", paramKey)
				}

				expectedQueryValues := tt.out[paramKey]

				for op, v := range values {
					if _, ok := expectedQueryValues[op]; !ok {
						t.Errorf("Values operator not exist: %s", op)
					}

					if !reflect.DeepEqual(expectedQueryValues[op], v) {
						t.Errorf("Values values not parsed correctly: expected len %v, got len %v", len(expectedQueryValues[op]), len(v))
					}
				}
			}
		})
	}
}
