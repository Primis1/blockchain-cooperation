package singleton

import "blockchain/pkg/logging"

var logs = logging.LoggingFacade


var Log = logs.NewLogger(logging.INFO)
var Err = logs.NewLogger(logging.ERR)

func idk() {
	Err.ErrorLog("fkdnfv")
}