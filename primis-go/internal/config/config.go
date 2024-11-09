package config

import (
	"blockchain/pkg/utils"
	"os"
)

func MustEnvironment() {
	err := os.Setenv("KEY_WORD", "primis-go")
	utils.HandleErr(err)

	os.Setenv("dbPath", "../tmp/blocks")
	os.Setenv("dbFile", "../tmp/blocks/MANIFEST")
	os.Setenv("genesisData", ".")
}
