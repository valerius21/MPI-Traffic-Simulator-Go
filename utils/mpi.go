package utils

import mpi "github.com/sbromberger/gompi"

var isMPI bool

func init() {
	isMPI = mpi.IsOn()
}

// IsMPI returns the isMPI variable
func IsMPI() bool {
	return isMPI
}
