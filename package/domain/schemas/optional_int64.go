package schemas

import "encoding/json"

// OptionalInt64 distinguishes "field absent" (Set=false) from an explicit null
// (Set=true, Value=nil), so API consumers can clear an override by sending null.
type OptionalInt64 struct {
	Set   bool
	Value *int64
}

// UnmarshalJSON implements json.Unmarshaler.
func (o *OptionalInt64) UnmarshalJSON(b []byte) error {
	o.Set = true
	if string(b) == "null" {
		o.Value = nil
		return nil
	}
	var v int64
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	o.Value = &v
	return nil
}

// MarshalJSON implements json.Marshaler.
func (o OptionalInt64) MarshalJSON() ([]byte, error) {
	if o.Value == nil {
		return []byte("null"), nil
	}
	return json.Marshal(*o.Value)
}
