package blockstate

import (
	"encoding/binary"
	"fmt"
	"hash/fnv"
	"sort"
)

// NetworkHash returns the network ID Bedrock uses for s when the server sends
// block IDs as hashes: FNV-1a 32 of the state as little-endian NBT,
// {name, states} with every compound's keys sorted. The version is not
// hashed. It fails for property values that are not byte, int32 or string.
func NetworkHash(s State) (uint32, error) {
	var b []byte
	b = append(b, tagCompound, 0, 0) // unnamed root compound
	b = appendString(append(b, tagString), "name")
	b = appendString(b, s.Name)
	b = appendString(append(b, tagCompound), "states")

	names := make([]string, 0, len(s.Properties))
	for n := range s.Properties {
		names = append(names, n)
	}
	sort.Strings(names)
	for _, n := range names {
		switch v := s.Properties[n].(type) {
		case uint8:
			b = append(appendString(append(b, tagByte), n), v)
		case int32:
			b = binary.LittleEndian.AppendUint32(appendString(append(b, tagInt), n), uint32(v)) //nolint:gosec // two's complement bytes are what NBT stores
		case string:
			b = appendString(appendString(append(b, tagString), n), v)
		default:
			return 0, fmt.Errorf("blockstate: hash %s: property %s has type %T", s.Name, n, v)
		}
	}
	b = append(b, tagEnd, tagEnd) // end of states, end of root

	h := fnv.New32a()
	_, _ = h.Write(b) // hash.Hash never returns an error
	return h.Sum32(), nil
}

const (
	tagEnd      = 0
	tagByte     = 1
	tagInt      = 3
	tagString   = 8
	tagCompound = 10
)

func appendString(b []byte, s string) []byte {
	b = binary.LittleEndian.AppendUint16(b, uint16(len(s))) //nolint:gosec // block names and properties are short
	return append(b, s...)
}
