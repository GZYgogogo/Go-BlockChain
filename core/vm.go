package core

type Instruction byte

const (
	InstrPushInt  Instruction = 0x0a
	InstrAdd      Instruction = 0x0b
	InstrPushByte Instruction = 0x0c
	InstrPack     Instruction = 0x0d
)

type Stack struct {
	data []any
	sp   int // stack pointer
}

func NewStack(size int) *Stack {
	return &Stack{
		data: make([]any, size),
		sp:   0,
	}
}

func (s *Stack) Push(val any) {
	s.data[s.sp] = val
	s.sp++
}

func (s *Stack) Pop() any {
	value := s.data[0]

	s.data = append(s.data[:0], s.data[1:]...)
	s.sp--

	return value
}

// VM used to execute the bytes from tx
type VM struct {
	data  []byte
	ip    int // instruction pointer
	stack *Stack
}

func NewVM(data []byte) *VM {
	return &VM{
		data:  data,
		ip:    0,
		stack: NewStack(1024),
	}
}

func (vm *VM) Run() error {
	for {
		instr := Instruction(vm.data[vm.ip])

		if err := vm.Exec(instr); err != nil {
			return err
		}

		vm.ip++

		if vm.ip > len(vm.data)-1 {
			break
		}
	}
	return nil
}

func (vm *VM) Exec(instr Instruction) error {
	switch instr {
	case InstrPushInt:
		vm.stack.Push(int(vm.data[vm.ip-1]))
	case InstrPushByte:
		vm.stack.Push(byte(vm.data[vm.ip-1]))
	case InstrPack:
		n := vm.stack.Pop().(int)
		arr := make([]byte, n)
		for i := 0; i < n; i++ {
			arr[i] = vm.stack.Pop().(byte)
		}
		vm.stack.Push(arr)
	case InstrAdd:
		a := vm.stack.Pop().(int)
		b := vm.stack.Pop().(int)
		vm.stack.Push(a + b)
	}
	return nil
}
