package agentflow

import (
	"reflect"
	"sort"
	"testing"
)

func TestState_CloneAndModify(t *testing.T) {
	original := NewState()
	original.Set("key1", "value1")
	original.Set("key2", 123)
	original.SetMeta("meta1", "mValue1")

	clone := original.Clone()

	// Modify clone
	clone.Set("key1", "newValue1")
	clone.Set("key3", true)
	clone.SetMeta("meta1", "newMetaValue1")
	clone.SetMeta("meta2", "mValue2")

	// Verify original is unchanged
	if val, ok := original.Get("key1"); !ok || val != "value1" {
		t.Errorf("Original data modified for key1: got %v, want %v", val, "value1")
	}
	if _, ok := original.Get("key3"); ok {
		t.Errorf("Original data unexpectedly contains key3")
	}
	if val, ok := original.GetMeta("meta1"); !ok || val != "mValue1" {
		t.Errorf("Original metadata modified for meta1: got %v, want %v", val, "mValue1")
	}
	if _, ok := original.GetMeta("meta2"); ok {
		t.Errorf("Original metadata unexpectedly contains meta2")
	}

	// Verify clone has changes
	if val, ok := clone.Get("key1"); !ok || val != "newValue1" {
		t.Errorf("Clone data not updated for key1: got %v, want %v", val, "newValue1")
	}
	if val, ok := clone.Get("key3"); !ok || val != true {
		t.Errorf("Clone data missing key3: want %v", true)
	}
	if val, ok := clone.GetMeta("meta1"); !ok || val != "newMetaValue1" {
		t.Errorf("Clone metadata not updated for meta1: got %v, want %v", val, "newMetaValue1")
	}
	if val, ok := clone.GetMeta("meta2"); !ok || val != "mValue2" {
		t.Errorf("Clone metadata missing meta2: want %v", "mValue2")
	}

	// Optional: Verify keys are as expected
	expectedOriginalKeys := []string{"key1", "key2"}
	actualOriginalKeys := original.Keys()
	sort.Strings(actualOriginalKeys) // Sort for consistent comparison
	if !reflect.DeepEqual(actualOriginalKeys, expectedOriginalKeys) {
		t.Errorf("Original keys mismatch: got %v, want %v", actualOriginalKeys, expectedOriginalKeys)
	}

	expectedCloneKeys := []string{"key1", "key2", "key3"}
	actualCloneKeys := clone.Keys()
	sort.Strings(actualCloneKeys) // Sort for consistent comparison
	if !reflect.DeepEqual(actualCloneKeys, expectedCloneKeys) {
		t.Errorf("Clone keys mismatch: got %v, want %v", actualCloneKeys, expectedCloneKeys)
	}
}

func TestState_GetSet(t *testing.T) {
	s := NewState()
	s.Set("string", "hello")
	s.Set("int", 42)
	s.Set("bool", true)

	if val, ok := s.Get("string"); !ok || val != "hello" {
		t.Errorf("Get string failed: got %v, %t", val, ok)
	}
	if val, ok := s.Get("int"); !ok || val != 42 {
		t.Errorf("Get int failed: got %v, %t", val, ok)
	}
	if val, ok := s.Get("bool"); !ok || val != true {
		t.Errorf("Get bool failed: got %v, %t", val, ok)
	}
	if _, ok := s.Get("missing"); ok {
		t.Errorf("Get missing unexpectedly succeeded")
	}
}
