package types

import (
	"encoding/json"
	"fmt"
)

func scanJSON[T any](v any, target *T) error {
	switch vv := v.(type) {
	case []byte:
		return json.Unmarshal(vv, target)
	case string:
		return json.Unmarshal([]byte(vv), target)
	default:
		return fmt.Errorf("unsupported type: %T", v)
	}
}
