package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type Data struct {
	Y            int
	ptr          *string
	structMember Impl
	ifMember     Interfc
	mapMember    map[string]interface{}
	sliceMember  []string
}

type Interfc interface {
	DoIt()
}

type Impl struct {
	implField map[string]string
}

func (i Impl) DoIt() {}

func TestValidateStructEmpty(t *testing.T) {
	t.Run("empty struct", func(t *testing.T) {
		var d Data
		err := ValidateStruct(d)
		if !assert.Error(t, err) {
			t.Fatalf("expected error, got nil")
		}
	})

	t.Run("nil pointer", func(t *testing.T) {
		var d *Data
		err := ValidateStruct(d)
		if !assert.Error(t, err) {
			t.Fatalf("expected error, got nil")
		}
	})

	t.Run("different field types", func(t *testing.T) {

		emptyString := ""

		d := Data{
			Y:   1,
			ptr: &emptyString,

			structMember: Impl{implField: make(map[string]string)},

			ifMember: Impl{},

			mapMember:   make(map[string]interface{}),
			sliceMember: make([]string, 0),
		}

		err := ValidateStruct(d)
		if !assert.NoError(t, err) {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("unititialized fields", func(t *testing.T) {
		var d struct {
			X Interfc
		}

		err := ValidateStruct(d)
		if !assert.Error(t, err) {
			t.Fatalf("expected error, got nil")
		}
	})

}
