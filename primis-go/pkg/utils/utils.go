package utils

import (
	"blockchain/pkg/logging"
	"bytes"
	"encoding/binary"
	"os"

	"github.com/mr-tron/base58"
)

var err = logging.Error

func HandleErr(T any) {
	if T != nil {
		err.Error(T)
	}
}

func ToHex(num int64) []byte {
	// NOTE binary buffer
	buff := new(bytes.Buffer)

	err := binary.Write(buff, binary.BigEndian, num)
	HandleErr(err)

	return buff.Bytes()
}

func Base58Encode(b []byte) []byte {
	encode := base58.Encode(b)

	return []byte(encode)
}

func Base58Decode(b []byte) []byte {
	decode, err := base58.Decode(string(b[:]))
	HandleErr(err)

	return decode
}

// NOTE fun fact: diff between base57 and base58 is that 58s misses {0, O, I, l, +, /}
// NOTE I.E base58 has protection from idiots, so user won't mess up with sending tokens 
// NOTE to "not that address"


func DirExist(dir string) bool {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return false
	}

	return true
}