package main

import (
	"fmt"
	"strconv"
)

// formatID converts an uint64 into a string representation.
func formatID(id uint64) string {
	return fmt.Sprintf("%03s", strconv.FormatUint(id*46649%6125, 36))
}

// IDGenerator can be used to generate unique IDs
type IDGenerator struct {
	index uint64
}

// REVIEW
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
	return formatID(g.UniqID())
}

var gen = MakeIDGenerator()

// UniqID returns and increments the current index
func UniqID() uint64 {
	return gen.UniqID()
}

// UniqIDf returns a new 'random' string id and increaes the index
func UniqIDf() string {
	return gen.UniqIDf()
}
