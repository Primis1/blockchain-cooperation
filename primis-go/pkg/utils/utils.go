package utils

import (
	"blockchain/pkg/logging"
	"bytes"
	"encoding/binary"
	"encoding/gob"

	"github.com/mr-tron/base58"
)

func HandleErr(T any) {
	if T != nil {
		logging.ErrorSingleTon().Error(T)
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

func GobEncoder(data interface{}) []byte {
	var buff bytes.Buffer

	enc := gob.NewEncoder(&buff)

	err := enc.Encode(data)
	HandleErr(err)
	return buff.Bytes()
}
