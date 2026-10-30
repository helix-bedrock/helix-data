package data

import (
	"bytes"
	_ "embed"
	"fmt"

	"github.com/helix-bedrock/protocol"
	"github.com/helix-bedrock/protocol/packet"
)

//go:embed 1.26.60/bds/packets/item_registry.bin
var itemRegistry []byte

// ItemRegistry returns every item of [Version] as BDS sends it to joining
// players in its ItemRegistry packet, in BDS's order: name, network ID,
// whether it is component based, version and component data. Each call
// decodes a fresh copy.
func ItemRegistry() (items []protocol.ItemEntry, err error) {
	defer func() {
		// The protocol reader stops on bad input by panicking; the data is
		// embedded and checked when written, so this only guards changes.
		if v := recover(); v != nil {
			items, err = nil, fmt.Errorf("data: item registry: %v", v)
		}
	}()
	var pk packet.ItemRegistry
	buf := bytes.NewBuffer(itemRegistry)
	pk.Marshal(protocol.NewReader(buf, 0, false))
	if buf.Len() != 0 {
		return nil, fmt.Errorf("data: item registry: %d bytes left", buf.Len())
	}
	return pk.Items, nil
}
