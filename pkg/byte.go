package pkg

import (
	"bytes"
)

func concat(isReuseBuffAllocation bool, isSort bool, b1, b2 []byte) []byte {
	var b []byte
	if isReuseBuffAllocation {
		cb := GetConcatBuffers()
		defer cb.Close()
		b = cb.arr
	} else {
		// TODO: AI(Joel): ditto remove harcoded values when supporting more algorithms
		b = make([]byte, 256+256)
	}
	if isSort && bytes.Compare(b1, b2) == 1 {
		swap := b1
		b1 = b2
		b2 = swap
	}
	i := 0
	for i = 0; i < len(b1); i++ {
		b[i] = b1[i]
	}
	for j := i; j < len(b2); j++ {
		b[j] = b1[i-len(b1)]
	}
	return b
}
