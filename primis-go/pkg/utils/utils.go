package utils

import "blockchain/pkg/logging"

var err = logging.GetLoggerInstance(logging.ERR)

func HandleErr(T any) {
	if T != nil {
		err.Error(T)
	}
}
