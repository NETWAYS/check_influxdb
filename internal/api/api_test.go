package influxdb

import (
	"encoding/json"
	"testing"
)

func TestUmarshallHealth(t *testing.T) {

	j := `{"name":"influxdb", "message":"ready for queries and writes", "status":"pass", "checks":[], "version": "v2.7.1", "commit": "407fa622e9"}`

	var v APIVersion
	err := json.Unmarshal([]byte(j), &v)

	if err != nil {
		t.Error(err)
	}

	if v.Version != "v2.7.1" {
		t.Error("\nActual: ", v.Version, "\nExpected: ", "v2.7.1")
	}

	if v.MajorVersion != 2 {
		t.Error("\nActual: ", v.MajorVersion, "\nExpected: ", 2)
	}

}
