package schemas

import (
	"encoding/json"
	"time"
)

// OptionalTime is a time.Time value that distinguishes "field absent" from
// an explicit null. JSON null decodes to Set=true, Value=nil; a missing field
// leaves Set=false. This allows API consumers to clear a time field by sending null.
type OptionalTime struct {
	Set   bool
	Value *time.Time
}

// UnmarshalJSON implements json.Unmarshaler.
func (o *OptionalTime) UnmarshalJSON(b []byte) error {
	o.Set = true
	if string(b) == "null" {
		o.Value = nil
		return nil
	}
	var t time.Time
	if err := json.Unmarshal(b, &t); err != nil {
		return err
	}
	o.Value = &t
	return nil
}

// MarshalJSON implements json.Marshaler.
func (o OptionalTime) MarshalJSON() ([]byte, error) {
	if o.Value == nil {
		return []byte("null"), nil
	}
	return json.Marshal(*o.Value)
}
