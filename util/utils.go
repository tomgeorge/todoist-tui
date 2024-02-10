// The dreaded util package
package util

import (
	"bytes"
	"encoding/json"
)

// Pretty print JSON data (not a struct)
// w/ special feature of not returning the error if it fails ;)
func PrettyJSON(buf []byte) string {
	var pretty bytes.Buffer
	_ = json.Indent(&pretty, buf, "", " ")
	return pretty.String()
}

// Pretty print a JSON representation of a struct
func PrettyStruct(v any) string {
	buf, _ := json.MarshalIndent(v, "", " ")
	return string(buf)
}
