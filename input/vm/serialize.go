package vm

import "encoding/binary"

const bytecodeSize = 1 + 8 + 8 // op + arg1 + arg2

var byteOrder = binary.LittleEndian

// SerializeProgram encodes a program's bytecodes.
func SerializeProgram(prog Program) []byte {
	var i int
	data := make([]byte, len(prog)*bytecodeSize)
	for _, bytecode := range prog {
		data[i] = byte(bytecode.op)
		byteOrder.PutUint64(data[i+1:], uint64(bytecode.arg1))
		byteOrder.PutUint64(data[i+9:], uint64(bytecode.arg2))
		i += bytecodeSize
	}
	return data
}

// DeserializeProgram decodes a program's bytecodes.
// It doesn't perform any validation, so it might panic
// or produce an invalid program if the input data is invalid.
func DeserializeProgram(data []byte) Program {
	var bc bytecode
	prog := make(Program, 0, len(data)/bytecodeSize)
	for i := 0; i < len(data); i += bytecodeSize {
		bc.op = op(data[i])
		bc.arg1 = int64(byteOrder.Uint64(data[i+1:]))
		bc.arg2 = int64(byteOrder.Uint64(data[i+9:]))
		prog = append(prog, bc)
	}
	return prog
}
