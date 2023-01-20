package bits

import "strconv"

// GetBit returns the bit at the given index in the given byte array
//
// The index is 0-based and the first bit is the most significant bit of the first byte
func GetBit(bytes *[]byte, bitIndex int) int {
	return int((*bytes)[bitIndex/8]&(1<<uint(7-bitIndex%8))) >> uint(7-bitIndex%8)
}

// SetBit sets the bit at the given index in the given byte array to the given value
//
// The index is 0-based and the first bit is the most significant bit of the first byte
func SetBit(bytes *[]byte, bitIndex int, value int) {
	if value == 1 {
		(*bytes)[bitIndex/8] |= 1 << uint(7-bitIndex%8)
	} else {
		(*bytes)[bitIndex/8] &= ^(1 << uint(7-bitIndex%8))
	}
}

func LeftShift(bytes *[]byte, n int) {
	maxBit := GetBit(bytes, 0)
	for i := 0; i < n; i++ {
		for j := 0; j < len(*bytes)*8-1; j++ {
			SetBit(bytes, j, GetBit(bytes, j+1))
		}
		SetBit(bytes, len(*bytes)*8-1, maxBit)
	}
}

func DisplayBits(bytes *[]byte) string {
	var result string
	for i := 0; i < len(*bytes)*8; i++ {
		result += strconv.Itoa(GetBit(bytes, i))
		if i%8 == 7 {
			result += " "
		}
	}
	return result
}
