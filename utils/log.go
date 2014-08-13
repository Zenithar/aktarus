package utils

import (
	"log"
	"os"
)

var logger *log.Logger

func init() {
	logger = log.New(os.Stdout, "[utils] ", log.LstdFlags)
}
