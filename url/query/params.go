package query

import (
	"net/url"
	"strings"
)

// Operators for operator -> values
// Empty operator means that no operator is found near the value.
// In this case server decides the default one
type Operators map[string][]string

// Values for parameter -> Operators
type Values map[string]Operators

// Parse parses query parameters for proper Values
// it can parse i.e. ?value=gt:10&name=in:name1,name22&date=lt:2018-04-01,gt:2018-01-01
func Parse(queryParams url.Values) Values {
	// prepare query map
	queryMap := make(Values, len(queryParams))

	// parse the map
	for k := range queryParams {
		queryValues := queryParams[k]
		// prepare operator queryValues map
		opValues := make(Operators)

		for _, qv := range queryValues {
			// check for queryValues
			val := strings.Split(qv, ",") // value separator [ie. val1,val2,val3]

			for _, v := range val {
				// split with operator
				op := strings.Split(v, ":") // operator separator [ie. asc:key1 or gt:10]
				currOp := Default           // empty mean no operator
				valueIndex := 0             // index of the actual query value, default to 0
				// check for operator
				if len(op) > 1 {
					// there is an operator
					valueIndex = 1 // value should be at index 1, because operators at 0
					currOp = op[0]
				}

				// assign value to operator
				opValues[currOp] = append(opValues[currOp], op[valueIndex])
			}

		}

		queryMap[k] = opValues
	}

	return queryMap
}
