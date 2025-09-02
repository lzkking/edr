package common

import "os"

var (
	Sig = make(chan os.Signal, 1)
)
