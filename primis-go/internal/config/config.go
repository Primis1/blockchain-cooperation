package config

import "os"

func MustEnvrionment() {
	err := os.Setenv("KEY_WORD", "primis-go")

	if err != nil {
		panic(err)
	}
}
