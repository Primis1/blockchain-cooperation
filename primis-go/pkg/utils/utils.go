package utils

import (
	"blockchain/pkg/logging"
	"bytes"
	"encoding/binary"
)

var err = logging.Err

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