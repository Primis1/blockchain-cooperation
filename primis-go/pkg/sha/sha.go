package sha

import (
	"encoding/binary"
	"fmt"
	"crypto/sha256"
)


// roots of first `24` prime number
var k = []uint{
	0x428a2f98, 0x71374491, 0xb5c0fbcf, 0xe9b5dba5, 0x3956c25b, 0x59f111f1, 0x923f82a4, 0xab1c5ed5,
	0xd807aa98, 0x12835b01, 0x243185be, 0x550c7dc3, 0x72be5d74, 0x80deb1fe, 0x9bdc06a7, 0xc19bf174,
	0xe49b69c1, 0xefbe4786, 0x0fc19dc6, 0x240ca1cc, 0x2de92c6f, 0x4a7484aa, 0x5cb0a9dc, 0x76f988da,
}

// initial hash values, defined by SHA standard
var h0 = []uint32{
	0x6a09e667,
	0xbb67ae85,
	0x3c6ef372,
	0xa54ff53a,
	0x510e527f,
	0x9b05688c,
	0x1f83d9ab,
	0x5be0cd19,
}

// right rotation on 32-bit integer
func rotateRight(x, n uint32) uint32 {
	return (x >> n) | (x << (32 - n))
}

// process a 512-bit chunk of data
func sha256Transform(hashValues []uint32, chunk []byte) {
	var w [64]uint32

	// Initialize first 16 words
	for i := 0; i < 16; i++ {
		w[i] = binary.BigEndian.Uint32(chunk[i*4 : (i+1)*4])
	}

	// Extend first 16 words into 48 words
	for i := 16; i < 64; i++ {
		s0 := rotateRight(w[i-15], 7) ^ rotateRight(w[i-15], 18) ^ (w[i-15] >> 3)
		s1 := rotateRight(w[i-2], 17) ^ rotateRight(w[i-2], 19) ^ (w[i-2] >> 10)
		w[i] = w[i-16] + s0 + w[i-7] + s1
	}

	// `8` 32-bit initial hash values (`H0`, `H7`)
	a, b, c, d, e, f, g, hVal := hashValues[0], hashValues[1], hashValues[2], hashValues[3], hashValues[4], hashValues[5], hashValues[6], hashValues[7]

	// For each 512-bit block, the algorithm iterates 64 times, updating hash values
	for i := 0; i < 64; i++ {
		s1 := rotateRight(e, 6) ^ rotateRight(e, 11) ^ rotateRight(e, 25)
		ch := (e & f) ^ (^e & g)
		
		// Ensure all components are of type uint32
		temp1 := hVal + s1 + ch + k[i] + w[i]
		s0 := rotateRight(a, 2) ^ rotateRight(a, 13) ^ rotateRight(a, 22) // Ensure s0 is computed here
		maj := (a & b) ^ (a & c) ^ (b & c)
		temp2 := s0 + maj

		// Update hash values
		hVal, g, f, e, d, c, b, a = g, f, e, d+temp1, c, b, a, temp1+temp2
	}

	// Hash to the result
	hashValues[0] += a
	hashValues[1] += b
	hashValues[2] += c
	hashValues[3] += d
	hashValues[4] += e
	hashValues[5] += f
	hashValues[6] += g
	hashValues[7] += hVal
}


func padMessage(message []byte) []byte {
	length := uint64(len(message) * 8)
	message = append(message, 0x80)

	// we add 0s untill we have required length
	for len(message)%64 != 56 {
		message = append(message, 0)
	}

	lenBytes := make([]byte, 0)

	binary.BigEndian.PutUint64(lenBytes, length)

	return append(message, lenBytes...)
}

func ComputeHash(data string) string {
	hash := make([]uint32, len(h0))

	// coping the values, but not reference to the array
	copy(hash, h0)

	paddedMessage := padMessage([]byte(data))

	for i := 0; i < len(paddedMessage); i += 64 {
		sha256Transform(hash, paddedMessage[i:i+64])
	}

	var result string

	for _, val := range hash {
		result += fmt.Sprintf("%08x", val)
	}

	return result
}
