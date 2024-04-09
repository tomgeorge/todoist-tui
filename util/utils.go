// The dreaded util package
package util

import (
	"bytes"
	"encoding/json"

	"github.com/tomgeorge/todoist-tui/model/drawer"
)

// Pretty print JSON data (not a struct)
// w/ special feature of not returning the error if it fails ;)
func PrettyJSON(buf []byte) string {
	var pretty bytes.Buffer
	_ = json.Indent(&pretty, buf, "", " ")
	return pretty.String()
}

// Pretty print a JSON representation of a struct
func PrettyMarshal(v any) ([]byte, error) {
	buf, err := json.MarshalIndent(v, "", " ")
	return buf, err
}

func AsPickable[E drawer.Pickable](items []E) []drawer.Pickable {
	pickables := make([]drawer.Pickable, len(items))
	for i, item := range items {
		pickables[i] = item
	}
	return pickables
}
