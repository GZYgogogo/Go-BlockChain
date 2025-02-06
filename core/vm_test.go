package core

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVM(t *testing.T) {
	data := []byte{0x03, 0x0a, 0x02, 0x0a, 0x0b, 0x4f, 0x0c, 0x4f, 0x0c, 0x46, 0x0c, 0x03, 0x0a, 0x0d, 0x0f}
	pushFoo := []byte{0x4f, 0x0c, 0x4f, 0x0c, 0x46, 0x0c, 0x03, 0x0a, 0x0d, 0xae}
	data = append(data, pushFoo...)
	constractState := NewState()
	vm := NewVM(data, constractState)
	assert.Nil(t, vm.Run())
	fmt.Printf("%+v\n", constractState)

	value := vm.stack.Pop().([]byte)
	// valueBytes, err := constractState.Get([]byte("FOO"))
	valueSerialized := deserializeInt64(value)
	assert.Equal(t, int64(2), valueSerialized)
}

func TestStack(t *testing.T) {
	stack := NewStack(1024)
	stack.Push(1)
	stack.Push(2)
	stack.Push(3)
	assert.Equal(t, 1, stack.Pop())
	fmt.Println(stack)
}

func TestMul(t *testing.T) {
	data := []byte{0x02, 0x0a, 0x02, 0x0a, 0xea}
	constractState := NewState()
	vm := NewVM(data, constractState)
	assert.Nil(t, vm.Run())
	result := vm.stack.Pop().(int)
	assert.Equal(t, 4, result)
}

func TestDiv(t *testing.T) {
	data := []byte{0x02, 0x0a, 0x04, 0x0a, 0xfd}
	constractState := NewState()
	vm := NewVM(data, constractState)
	assert.Nil(t, vm.Run())
	result := vm.stack.Pop().(int)
	assert.Equal(t, 2, result)
}
