package agentflow

import (
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
	originalData := original.GetData()
	if val, _ := originalData["key1"]; val != "value1" {
		t.Errorf("Original data modified for key1: got %v, want %v", val, "value1")
	}
	if _, ok := originalData["key3"]; ok {
		t.Errorf("Original data unexpectedly contains key3")
	}
	originalMeta := original.GetMetadata()
	if val, _ := originalMeta["meta1"]; val != "mValue1" {
		t.Errorf("Original metadata modified for meta1: got %v, want %v", val, "mValue1")
	}
	if _, ok := originalMeta["meta2"]; ok {
		t.Errorf("Original metadata unexpectedly contains meta2")
	}

	// Verify clone has changes
	cloneData := clone.GetData()
	if val, _ := cloneData["key1"]; val != "newValue1" {
		t.Errorf("Clone data not updated for key1: got %v, want %v", val, "newValue1")
	}
	if val, _ := cloneData["key3"]; val != true {
		t.Errorf("Clone data missing key3: want %v", true)
	}
	cloneMeta := clone.GetMetadata()
	if val, _ := cloneMeta["meta1"]; val != "newMetaValue1" {
		t.Errorf("Clone metadata not updated for meta1: got %v, want %v", val, "newMetaValue1")
	}
	if val, _ := cloneMeta["meta2"]; val != "mValue2" {
		t.Errorf("Clone metadata missing meta2: want %v", "mValue2")
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
