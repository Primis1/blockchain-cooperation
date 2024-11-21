package config

import (
	"blockchain/pkg/utils"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
)

var (
	dbPath  = "/tmp/blocks/"
	wallets = "/tmp/wallet_%s.data"
	dbFile  = "/tmp/blocks/MANIFEST"
)

func trimAndCombine(s string) string {
	_, path, _, ok := runtime.Caller(2)
	if !ok {
		utils.DisplayErr("caller does not work")
	}

	r, _ := regexp.Compile(`/?main\.go/?`)
	path = fmt.Sprintf("%s%s", r.ReplaceAllString(path, ""), s)

	return path
}

func MustEnvironment() {
	dbPath = trimAndCombine(dbPath)
	wallets = trimAndCombine(wallets)
	dbFile = trimAndCombine(dbFile)

	err := os.Setenv("KEY_WORD", "bc")
	utils.DisplayErr(err)

	err = os.Setenv("dbPath", dbPath)
	utils.DisplayErr(err)

	// default node id
	err = os.Setenv("NODE_ID", "3000")
	utils.DisplayErr(err)

	err = os.Setenv("wallets", wallets)
	utils.DisplayErr(err)

	err = os.Setenv("dbFile", dbFile)
	utils.DisplayErr(err)

	err = os.Setenv("genesisData", ".")
	utils.DisplayErr(err)
}

func CreateFilesIfNotExist() {
	if _, err := os.Stat(dbPath); errors.Is(err, os.ErrNotExist) {
		if err := os.MkdirAll(filepath.Dir(dbPath), 0770); err != nil {
			utils.DisplayErr("folder can't be created")
		}
		os.Create(dbPath)
	}
}
