package config

import "os"

func MustEnvironment() {
	err := os.Setenv("KEY_WORD", "primis-go")

	if err != nil {
		panic(err)
	}
}
