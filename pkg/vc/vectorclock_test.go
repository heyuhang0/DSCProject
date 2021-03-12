package vc

import (
	"reflect"
	"testing"
)

func TestFib(t *testing.T) {
	var (
		in       = 7
		expected = 13
	)
	actual := Fib(in)
	if actual != expected {
		t.Errorf("Fib(%d) = %d; expected %d", in, actual, expected)
	}
}

func TestNewVectorClock(t *testing.T) {
	var (
		in       = 7
		expected = vclock{7: 0}
	)
	actual := NewVectorClock(in)
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("NewVectorClock(%v) = %v; expected %v", in, actual, expected)
	}
}

func TestVclock_MergeClock(t *testing.T) {
	// prepare two vector clocks
	c1 := NewVectorClock(1)
	c2 := NewVectorClock(2)
	c1.Advance(1)
	c1.Advance(1)
	c1.Advance(1)
	c2.Advance(2)

	expected := vclock{1: 4, 2: 1}

	c1.MergeClock(1, c2)
	if !reflect.DeepEqual(c1, expected) {
		t.Errorf("MergeClock(%v) = %v; expected %v", c2, c1, expected)
	}
}

func TestVclock_MergeClock2(t *testing.T) {
	// Test detection of causality violation
	// prepare two vector clocks
	c1 := vclock{1: 3, 2: 5} // local clock
	c2 := vclock{1: 2, 2: 5, 3: 4}
	expected := true

	actual := c1.MergeClock(1, c2)
	if actual != expected {
		t.Errorf("MergeClock(%v) = %v; expected %v", c2, actual, expected)
	}
}

func TestVclock_Advance(t *testing.T) {
	c1 := vclock{1: 3, 2: 5}

	expected := vclock{1: 4, 2: 5}
	c1.Advance(1)
	if !reflect.DeepEqual(c1, expected) {
		t.Errorf("MergeClock(%v) = %v; expected %v", c1, c1, expected)
	}
}
