package sha

import (
	"encoding/binary"
)

// roots of first `64` prime number
var k = []uint32{
	0x428a2f98, 0x71374491, 0xb5c0fbcf, 0xe9b5dba5, 0x3956c25b, 0x59f111f1, 0x923f82a4, 0xab1c5ed5,
	0xd807aa98, 0x12835b01, 0x243185be, 0x550c7dc3, 0x72be5d74, 0x80deb1fe, 0x9bdc06a7, 0xc19bf174,
	0xe49b69c1, 0xefbe4786, 0x0fc19dc6, 0x240ca1cc, 0x2de92c6f, 0x4a7484aa, 0x5cb0a9dc, 0x76f988da,
	0x983e5152, 0xa831c66d, 0xb00327c8, 0xbf597fc7, 0xc6e00bf3, 0xd5a79147, 0x06ca6351, 0x14292967,
	0x27b70a85, 0x2e1b2138, 0x4d2c6dfc, 0x53380d13, 0x650a7354, 0x766a0abb, 0x81c2c92e, 0x92722c85,
	0xa2bfe8a1, 0xa81a664b, 0xc24b8b70, 0xc76c51a3, 0xd192e819, 0xd6990624, 0xf40e3585, 0x106aa070,
	0x19b4f63b, 0x1e376c4f, 0x2748774c, 0x34b0bcb5, 0x391c0cb3, 0x4ed8aa11, 0x5b9cca4f, 0x682e6ff3,
	0x748f82ee, 0x78a5636f, 0x84c87814, 0x8cc70208, 0x90befffa, 0xa4506ceb, 0xbef9a3f7, 0xc67178f2,
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

	// We add 0s until we have required length
	for len(message)%64 != 56 {
		message = append(message, 0)
	}

	// Allocate a slice with a size of 8 bytes to hold the length
	lenBytes := make([]byte, 8) // Create a byte slice of length 8
	binary.BigEndian.PutUint64(lenBytes, length)

	return append(message, lenBytes...)
}

func ComputeHash(data []byte) [32]byte {
	hash := make([]uint32, len(h0))

	// Copying the values, but not reference to the array
	copy(hash, h0)

	paddedMessage := padMessage(data)

	for i := 0; i < len(paddedMessage); i += 64 {
		sha256Transform(hash, paddedMessage[i:i+64])
	}

	// Construct the final hash output as a byte slice
	result := [32]byte{} // SHA-256 produces a 32-byte hash
	for i, val := range hash {
		binary.BigEndian.PutUint32(result[i*4:], val) // Correctly convert uint32 to bytes
	}

	return result
}
