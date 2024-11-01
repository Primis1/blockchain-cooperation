package singleton

import "blockchain/pkg/logging"

type dependency struct {
	Logging logging.Application
}

const sing = &dependency{}
