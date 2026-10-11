package main

import (
	"math/rand/v2"
	"os"
)

const ProgOffset = 0x200
const SpriteOffset = 0x50

type Chip struct {
	memory  [4096]uint8 // where the program and sprites goes
	V       [16]uint8   // V registers
	index   uint16
	pc      uint16          // program counter
	stack   [16]uint16      //opcode stack
	sp      uint8           //stack pointer
	dt, st  uint8           //delay and sound timers
	keypad  [16]uint8       // controller
	display [64 * 32]uint32 // the pixels that will be displayed
	opcode  uint16          // the operation
}

func (chip *Chip) loadROM(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Read(chip.memory[ProgOffset:])
	if err != nil {
		return err
	}
	println("ROM loaded:", path)
	chip.pc = ProgOffset
	return nil
}

func (chip *Chip) loadSprites() {
	var fontset = [80]uint8{
		0xF0, 0x90, 0x90, 0x90, 0xF0, // 0
		0x20, 0x60, 0x20, 0x20, 0x70, // 1
		0xF0, 0x10, 0xF0, 0x80, 0xF0, // 2
		0xF0, 0x10, 0xF0, 0x10, 0xF0, // 3
		0x90, 0x90, 0xF0, 0x10, 0x10, // 4
		0xF0, 0x80, 0xF0, 0x10, 0xF0, // 5
		0xF0, 0x80, 0xF0, 0x90, 0xF0, // 6
		0xF0, 0x10, 0x20, 0x40, 0x40, // 7
		0xF0, 0x90, 0xF0, 0x90, 0xF0, // 8
		0xF0, 0x90, 0xF0, 0x10, 0xF0, // 9
		0xF0, 0x90, 0xF0, 0x90, 0x90, // A
		0xE0, 0x90, 0xE0, 0x90, 0xE0, // B
		0xF0, 0x80, 0x80, 0x80, 0xF0, // C
		0xE0, 0x90, 0x90, 0x90, 0xE0, // D
		0xF0, 0x80, 0xF0, 0x80, 0xF0, // E
		0xF0, 0x80, 0xF0, 0x80, 0x80, // F
	}
	copy(chip.memory[SpriteOffset:], fontset[:])
}

//func (chip *Chip) OP_0NNN() {
//
//} // SYS addr (ignored)

func (chip *Chip) OP_00E0() {
	clear(chip.display[:])
} // CLS

func (chip *Chip) OP_00EE() {
	chip.sp--
	chip.pc = chip.stack[chip.sp]
} // RET

func (chip *Chip) OP_1NNN() {
	chip.pc = chip.NNN()
} // JP addr

func (chip *Chip) OP_2NNN() {
	chip.stack[chip.sp] = chip.pc
	chip.sp++
	chip.pc = chip.NNN()
} // CALL addr

func (chip *Chip) OP_3XNN() {
	if chip.V[chip.X()] == chip.NN() {
		chip.pc += 2
	}
} // SE Vx, byte

func (chip *Chip) OP_4XNN() {
	if chip.V[chip.X()] != chip.NN() {
		chip.pc += 2
	}
} // SNE Vx, byte

func (chip *Chip) OP_5XY0() {
	if chip.V[chip.X()] == chip.V[chip.Y()] {
		chip.pc += 2
	}
} // SE Vx, Vy

func (chip *Chip) OP_6XNN() {
	chip.V[chip.X()] = chip.NN()
} // LD Vx, byte

func (chip *Chip) OP_7XNN() {
	chip.V[chip.X()] += chip.NN()
} // ADD Vx, byte

func (chip *Chip) OP_8XY0() {
	chip.V[chip.X()] = chip.V[chip.Y()]
} // LD Vx, Vy

func (chip *Chip) OP_8XY1() {
	x := chip.X()
	chip.V[x] = chip.V[x] | chip.V[chip.Y()]
} // OR Vx, Vy

func (chip *Chip) OP_8XY2() {
	x := chip.X()
	chip.V[x] = chip.V[x] & chip.V[chip.Y()]
} // AND Vx, Vy

func (chip *Chip) OP_8XY3() {
	x := chip.X()
	chip.V[x] = chip.V[x] ^ chip.V[chip.Y()]
} // XOR Vx, Vy

func (chip *Chip) OP_8XY4() {
	x := chip.X()
	y := chip.Y()

	vx := chip.V[x]
	vy := chip.V[y]

	sum := uint16(vx) + uint16(vy)

	if sum > 0xFF {
		chip.V[0xF] = 1
	} else {
		chip.V[0xF] = 0
	}
	chip.V[x] = uint8(sum)
} // ADD Vx, Vy

func (chip *Chip) OP_8XY5() {
	x := chip.X()
	y := chip.Y()

	vx := chip.V[x]
	vy := chip.V[y]

	if vx > vy {
		chip.V[0xF] = 1
	} else {
		chip.V[0xF] = 0
	}
	chip.V[x] -= chip.V[y]
} // SUB Vx, Vy

func (chip *Chip) OP_8XY6() {
	x := chip.X()
	chip.V[0xF] = chip.V[x] & 1
	chip.V[x] = chip.V[x] >> 1
} // SHR Vx

func (chip *Chip) OP_8XY7() {
	x := chip.X()
	y := chip.Y()

	if chip.V[y] > chip.V[x] {
		chip.V[0xF] = 1
	} else {
		chip.V[0xF] = 0
	}
	chip.V[x] = chip.V[y] - chip.V[x]
} // SUBN Vx, Vy

func (chip *Chip) OP_8XYE() {
	x := chip.X()
	chip.V[0xF] = (chip.V[x] >> 7) & 1
	chip.V[x] = chip.V[x] << 1
} // SHL Vx

func (chip *Chip) OP_9XY0() {
	x := chip.X()
	y := chip.Y()
	if chip.V[y] != chip.V[x] {
		chip.pc += 2
	}
} // SNE Vx, Vy

func (chip *Chip) OP_ANNN() {
	chip.index = chip.NNN()
} // LD I, addr

func (chip *Chip) OP_BNNN() {
	chip.pc = chip.NNN() + uint16(chip.V[0])
} // JP V0, addr

func (chip *Chip) OP_CXNN() {
	r := uint8(rand.IntN(256))
	chip.V[chip.X()] = r & chip.NN()
} // RND Vx, byte

func (chip *Chip) OP_DXYN() {
	x := chip.V[chip.X()] % 64
	y := chip.V[chip.Y()] % 32
	height := chip.N()

	chip.V[0xF] = 0 // reset collision flag

	for row := range height {
		spriteByte := chip.memory[chip.index+uint16(row)]

		for col := range 8 {
			pixel := (spriteByte >> (7 - col)) & 1

			px := (x + uint8(col)) % 64
			py := (y + row) % 32

			idx := int(py)*64 + int(px)

			if pixel == 1 {
				if chip.display[idx] == 1 {
					chip.V[0xF] = 1 // collision
				}
				chip.display[idx] ^= 1 // XOR draw
			}
		}
	}
}

// DRW Vx, Vy, nibble

func (chip *Chip) pressed(key uint8) bool {
	if int(key) >= len(chip.keypad) {
		return false
	}
	return chip.keypad[key] != 0
}

func (chip *Chip) OP_EX9E() {
	if chip.pressed(chip.V[chip.X()]) {
		chip.pc += 2
	}
} // SKP Vx

func (chip *Chip) OP_EXA1() {
	if !chip.pressed(chip.V[chip.X()]) {
		chip.pc += 2
	}
} // SKNP Vx

func (chip *Chip) OP_FX07() {
	chip.V[chip.X()] = chip.dt
} // LD Vx, DT

func (chip *Chip) OP_FX0A() {
	x := chip.X()

	for key := uint8(0); key < 16; key++ {
		if chip.pressed(key) {
			chip.V[x] = key
			return
		}
	}
	// No key is down, so run this instruction again next cycle.
	chip.pc -= 2
} // LD Vx, K

func (chip *Chip) OP_FX15() {
	chip.dt = chip.V[chip.X()]
} // LD DT, Vx

func (chip *Chip) OP_FX18() {
	chip.st = chip.V[chip.X()]
} // LD ST, Vx

func (chip *Chip) OP_FX1E() {
	chip.index += uint16(chip.V[chip.X()])
} // ADD I, Vx

func (chip *Chip) OP_FX29() {
	digit := chip.V[chip.X()]
	chip.index = uint16(SpriteOffset + digit*5)
} // LD F, Vx

func (chip *Chip) OP_FX33() {
	value := chip.V[chip.X()]
	chip.memory[chip.index] = value / 100
	chip.memory[chip.index+1] = (value / 10) % 10
	chip.memory[chip.index+2] = value % 10
} // LD B, Vx

func (chip *Chip) OP_FX55() {
	x := chip.X()
	for i := uint8(0); i <= x; i++ {
		chip.memory[chip.index+uint16(i)] = chip.V[i]
	}
} // LD [I], Vx

func (chip *Chip) OP_FX65() {
	x := chip.X()
	for i := uint8(0); i <= x; i++ {
		chip.V[i] = chip.memory[chip.index+uint16(i)]
	}
} // LD Vx, [I]

func (chip *Chip) fetch() {
	hi := chip.memory[chip.pc]
	low := chip.memory[chip.pc+1]
	chip.opcode = uint16(hi)<<8 | uint16(low)
	chip.pc += 2
}

func (chip *Chip) decode() {

	switch chip.opcode & 0xF000 {
	case 0x0000:
		switch chip.opcode & 0x00FF {
		case 0x00E0:
			chip.OP_00E0()
		case 0x00EE:
			chip.OP_00EE()
		}
	case 0x1000:
		chip.OP_1NNN()
	case 0x2000:
		chip.OP_2NNN()
	case 0x3000:
		chip.OP_3XNN()
	case 0x4000:
		chip.OP_4XNN()
	case 0x5000:
		chip.OP_5XY0()
	case 0x6000:
		chip.OP_6XNN()
	case 0x7000:
		chip.OP_7XNN()
	case 0x8000:
		switch chip.opcode & 0x000F {
		case 0x0000:
			chip.OP_8XY0()
		case 0x0001:
			chip.OP_8XY1()
		case 0x0002:
			chip.OP_8XY2()
		case 0x0003:
			chip.OP_8XY3()
		case 0x0004:
			chip.OP_8XY4()
		case 0x0005:
			chip.OP_8XY5()
		case 0x0006:
			chip.OP_8XY6()
		case 0x0007:
			chip.OP_8XY7()
		case 0x000E:
			chip.OP_8XYE()
		}
	case 0x9000:
		chip.OP_9XY0()
	case 0xA000:
		chip.OP_ANNN()
	case 0xB000:
		chip.OP_BNNN()
	case 0xC000:
		chip.OP_CXNN()
	case 0xD000:
		chip.OP_DXYN()
	case 0xE000:
		switch chip.opcode & 0x00FF {
		case 0x009E:
			chip.OP_EX9E()
		case 0x00A1:
			chip.OP_EXA1()
		}
	case 0xF000:
		switch chip.opcode & 0x00FF {
		case 0x0007:
			chip.OP_FX07()
		case 0x000A:
			chip.OP_FX0A()
		case 0x0015:
			chip.OP_FX15()
		case 0x0018:
			chip.OP_FX18()
		case 0x001E:
			chip.OP_FX1E()
		case 0x0029:
			chip.OP_FX29()
		case 0x0033:
			chip.OP_FX33()
		case 0x0055:
			chip.OP_FX55()
		case 0x0065:
			chip.OP_FX65()
		}
	}
}

func (chip *Chip) execute() {
	chip.fetch()
	chip.decode()
}

func (chip *Chip) cycle() {
	chip.execute()
}

func (chip *Chip) tickTimers() {
	if chip.dt > 0 {
		chip.dt--
	}
	if chip.st > 0 {
		chip.st--
	}
}

func (chip *Chip) X() uint8 {
	return uint8((chip.opcode & 0x0F00) >> 8)
}

func (chip *Chip) Y() uint8 {
	return uint8((chip.opcode & 0x00F0) >> 4)
}

func (chip *Chip) NN() uint8 {
	return uint8(chip.opcode & 0x00FF)
}

func (chip *Chip) NNN() uint16 {
	return chip.opcode & 0x0FFF
}

func (chip *Chip) N() uint8 {
	return uint8(chip.opcode & 0x000F)
}
