package model

import "testing"

func TestEventManager_Register(t *testing.T) {
	m := NewEventManager()
	e := NewEvent("handler name")

	if m.Contains(e) {
		t.Fatal("m should not contain e yet")
	}
	m.Register(e)
	if !m.Contains(e) {
		t.Fatal("m should contain e")
	}
}

func TestEventManager_Unregister(t *testing.T) {
	m := NewEventManager()
	e := NewEvent("handler name")
	m.Register(e)

	if !m.Contains(e) {
		t.Fatal("m should contain e")
	}
	m.Unregister(e)
	if m.Contains(e) {
		t.Fatal("m should not contain e anymore")
	}

	// Should be idempotent
	m.Unregister(e)
	if m.Contains(e) {
		t.Fatal("m should not contain e anymore")
	}
}
