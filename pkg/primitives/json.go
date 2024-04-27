package primitives

import (
	"fmt"

	jsoniter "github.com/json-iterator/go"
)

func MarshalJSON(v any) ([]byte, error) {
	data, err := jsoniter.ConfigCompatibleWithStandardLibrary.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("marshal error %v: %w", v, err)
	}

	return data, nil
}

func UnmarshalJSON(data []byte, v any) error {
	err := jsoniter.ConfigCompatibleWithStandardLibrary.Unmarshal(data, v)
	if err != nil {
		return fmt.Errorf("unmarshal error %s: %w", data, err)
	}

	return nil
}
