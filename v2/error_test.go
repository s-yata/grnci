package grnci

import (
	"log"
	"testing"
)

func TestNewError(t *testing.T) {
	data := map[string]interface{}{
		"string": "value",
		"int":    100,
	}
	err := NewError(AddressError, data)
	if err.Code != AddressError {
		t.Fatalf("NewError failed: Code: actual = %d, want = %d", err.Code, AddressError)
	}
	for k, v := range data {
		if err.Data[k] != v {
			t.Fatalf("NewError failed: Data[\"%s\"]: actual = %s, want = %s", k, err.Data[k], v)
		}
	}
	log.Printf("err = %v", err)
}

func TestEnhanceError(t *testing.T) {
	data := map[string]interface{}{
		"string": "value",
		"int":    100,
	}
	newData := map[string]interface{}{
		"string": "value2",
		"int":    1000,
		"float":  1.0,
	}
	err := NewError(AddressError, data)
	err = EnhanceError(err, newData)
	if err.Code != AddressError {
		t.Fatalf("NewError failed: Code: actual = %d, want = %d", err.Code, AddressError)
	}
	for k, v := range newData {
		if err.Data[k] != v {
			t.Fatalf("NewError failed: Data[\"%s\"]: actual = %s, want = %s", k, err.Data[k], v)
		}
	}
}
