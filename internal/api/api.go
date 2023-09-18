package influxdb

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// APIVersion is mainly for identifying the API Version of InfluxDB
type APIVersion struct {
	Version      string `json:"version"`
	MajorVersion int
}

// Custom Unmarshal since we might want to add or parse
// further fields in the future. This is simpler to extend and
// to test here than during the CheckPlugin logic.
func (v *APIVersion) UnmarshalJSON(b []byte) error {
	type Temp APIVersion

	t := (*Temp)(v)

	if err := json.Unmarshal(b, t); err != nil {
		return err
	}

	// Could also use some semver package,
	// but decided against the dependency
	if v.Version != "" {
		version := strings.Split(strings.TrimLeft(v.Version, "v"), ".")
		majorVersion, convErr := strconv.Atoi(version[0])

		if convErr != nil {
			return fmt.Errorf("could not determine version")
		}

		v.MajorVersion = majorVersion
	}

	return nil
}
