package core

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVM(t *testing.T) {
	contractStatus := NewStatus()
	// push F to stack
	// push O to stack
	// push O to stack
	// pack FOO
	// data := []byte{0x03, 0x0a, 0x46, 0x0c, 0x4f, 0x0c, 0x4f, 0x0c, 0x0d}
	data := []byte{0x03, 0x0a, 0x46, 0x0c, 0x4f, 0x0c, 0x4f, 0x0c, 0x0d, 0x05, 0x0a, 0x0f}
	vm := NewVM(data, contractStatus)
	assert.Nil(t, vm.Run())
	fmt.Printf("%+v", vm.stack.data)
	fmt.Printf("%+v", vm.contractStatus)
	valueBytes, err := vm.contractStatus.Get("FOO")
	value := deSerializeInt64(valueBytes)
	assert.Nil(t, err)
	assert.Equal(t, value, int64(5))
	// assert.Equal(t, 4, len(vm.stack.data))
}

func TestStack(t *testing.T) {
	stack := NewStack(1024)
	stack.Push(1)
	stack.Push(2)
	stack.Push(3)
	assert.Equal(t, 1, stack.Pop())
	fmt.Println(stack)
}
