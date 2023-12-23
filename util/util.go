package util

// IsBitSet checks if the bit at the specified position is set in a byte.
func IsBitSet(n byte, pos uint) bool {
	return (n & (1 << pos)) != 0
}

// ToggleBit toggles the bit at the specified position in a byte.
func ToggleBit(n byte, pos uint) byte {
	return n ^ (1 << pos)
}

// Complement complements all bits in a byte.
func Complement(num byte) byte {
	return num ^ 0xFF
}

// UnsetBit unsets the bit at the specified position in a byte.
func UnsetBit(n byte, pos uint) byte {
	return n &^ (1 << pos)
}

// SetBit sets the bit at the specified position in a byte.
func SetBit(n byte, pos uint) byte {
	return n | (1 << pos)
}

// IsAllBitsZero checks if all bits in a byte slice are zero.
func IsAllBitsZero(s []byte) bool {
	for _, v := range s {
		if v != 0 {
			return false
		}
	}
	return true
}
