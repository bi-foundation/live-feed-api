package events

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSubscriptionStack_Push(t *testing.T) {
	stack := NewSubscriptionStack(nil)
	e1 := []byte("1")
	e2 := []byte("2")

	stack.Push(e1)
	stack.Push(e2)

	s := stack.(*subscriptionStack)

	assert.Contains(t, s.events, e1)
	assert.Contains(t, s.events, e2)

	// the event is pushed to the top of the stack, so pop should return the last event first
	_, p2 := stack.Pop()
	assert.Equal(t, e2, p2)

	_, p1 := stack.Pop()
	assert.Equal(t, e1, p1)
}

func TestSubscriptionStack_Add(t *testing.T) {
	stack := NewSubscriptionStack(nil)
	e1 := []byte("1")
	e2 := []byte("2")

	stack.Add(e1)
	stack.Add(e2)

	s := stack.(*subscriptionStack)

	assert.Contains(t, s.events, e1)
	assert.Contains(t, s.events, e2)

	// the event is add to the back of the stack, so pop should return the first event that was added
	_, p1 := stack.Pop()
	assert.Equal(t, e1, p1)

	_, p2 := stack.Pop()
	assert.Equal(t, e2, p2)
}

func TestSubscriptionStack_IsProcessing(t *testing.T) {
	stack := NewSubscriptionStack(nil)

	assert.Equal(t, false, stack.IsProcessing())
	stack.Processing(true)
	assert.Equal(t, true, stack.IsProcessing())
	stack.Processing(false)
	assert.Equal(t, false, stack.IsProcessing())
}
