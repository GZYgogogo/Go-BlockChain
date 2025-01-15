package core

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVM(t *testing.T) {
	data := []byte{0x03, 0x0a, 0x31, 0x0c, 0x40, 0x0c, 0x40, 0x0c, 0x0d}
	vm := NewVM(data)
	assert.Nil(t, vm.Run())
	fmt.Printf("%+v", string(vm.stack.Pop().([]byte)))
	assert.Equal(t, 4, len(vm.stack.data))
}

func TestStack(t *testing.T) {
	stack := NewStack(1024)
	stack.Push(1)
	stack.Push(2)
	stack.Push(3)
	assert.Equal(t, 1, stack.Pop())
	fmt.Println(stack)
}
