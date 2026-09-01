package database

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

type JSONMap map[string]interface{}

func (m *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*m = JSONMap{}
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to scan JSONMap: unexpected type %T", value)
	}
	if len(bytes) == 0 {
		*m = JSONMap{}
		return nil
	}

	out := make(JSONMap)
	if err := json.Unmarshal(bytes, &out); err != nil {
		return err
	}
	*m = out
	return nil
}

func (m JSONMap) Value() (driver.Value, error) {
	if m == nil {
		return []byte("{}"), nil
	}
	return json.Marshal(m)
}

func (m JSONMap) ToMap() map[string]interface{} {
	if m == nil {
		return map[string]interface{}{}
	}
	return map[string]interface{}(m)
}
