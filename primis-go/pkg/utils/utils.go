package utils

import "blockchain/pkg/logging"

var err = logging.Err

func HandleErr(T any) {
	if T != nil {
		err.Error(T)
	}
}
