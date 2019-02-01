package utils

import (
	"fmt"
	"strconv"
)

// FormatID converts an uint64 into a string representation.
func FormatID(id uint64) string {
	return fmt.Sprintf("%03s", strconv.FormatUint(id*46649%6125, 36))
}

// IDGenerator can be used to generate unique IDs
type IDGenerator struct {
	index uint64
}

// MakeIDGenerator makes an IDGenerator
func MakeIDGenerator() IDGenerator {
	return IDGenerator{
		index: 0,
	}
}

// UniqID returns and increments the current index
func (g *IDGenerator) UniqID() uint64 {
	res := g.index
	g.index++
	return res
}

// UniqIDf returns a new 'random' string id and increaes the index
func (g *IDGenerator) UniqIDf() string {
	return FormatID(g.UniqID())
}

// Global is the global ID generator
var Global = MakeIDGenerator()
