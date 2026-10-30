package data

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"errors"
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
func ItemRegistry() ([]protocol.ItemEntry, error) {
	var pk packet.ItemRegistry
	if err := decodePacket(&pk, itemRegistry, 0); err != nil {
		return nil, fmt.Errorf("data: item registry: %w", err)
	}
	return pk.Items, nil
}

//go:embed 1.26.60/bds/packets/biome_definition_list.bin
var biomeDefinitions []byte

// BiomeDefinitions returns the BiomeDefinitionList packet BDS sends joining
// players: every biome with its ID, climate, tags and the data clients use
// to generate chunks themselves, and the string list those refer to. Each
// call decodes a fresh copy.
func BiomeDefinitions() (*packet.BiomeDefinitionList, error) {
	var pk packet.BiomeDefinitionList
	if err := decodePacket(&pk, biomeDefinitions, 0); err != nil {
		return nil, fmt.Errorf("data: biome definitions: %w", err)
	}
	return &pk, nil
}

//go:embed 1.26.60/bds/biome_ids.json
var biomeIDs []byte

// BiomeIDs returns the numeric ID of every vanilla biome of [Version], the
// ID chunks store, as BDS registers them. Each call decodes a fresh copy.
func BiomeIDs() (map[string]int16, error) {
	var doc struct {
		Biomes map[string]int16 `json:"biomes"`
	}
	if err := json.Unmarshal(biomeIDs, &doc); err != nil {
		return nil, fmt.Errorf("data: biome IDs: %w", err)
	}
	return doc.Biomes, nil
}

//go:embed 1.26.60/bds/packets/available_actor_identifiers.bin
var actorIdentifiers []byte

// ActorIdentifiers returns the AvailableActorIdentifiers packet BDS sends
// joining players: every entity type the client may see. Each call decodes
// a fresh copy.
func ActorIdentifiers() (*packet.AvailableActorIdentifiers, error) {
	var pk packet.AvailableActorIdentifiers
	if err := decodePacket(&pk, actorIdentifiers, 0); err != nil {
		return nil, fmt.Errorf("data: actor identifiers: %w", err)
	}
	return &pk, nil
}

//go:embed 1.26.60/bds/packets/creative_content.bin
var creativeContent []byte

// CreativeContent returns the CreativeContent packet BDS sends joining
// players: the creative inventory's groups and items. Each call decodes a
// fresh copy.
func CreativeContent() (*packet.CreativeContent, error) {
	shield, err := shieldID()
	if err != nil {
		return nil, err
	}
	var pk packet.CreativeContent
	if err := decodePacket(&pk, creativeContent, shield); err != nil {
		return nil, fmt.Errorf("data: creative content: %w", err)
	}
	return &pk, nil
}

//go:embed 1.26.60/bds/packets/crafting_data.bin
var craftingData []byte

// CraftingData returns the CraftingData packet BDS sends joining players:
// every recipe, potion and container mix, and material reducer. Each call
// decodes a fresh copy.
func CraftingData() (*packet.CraftingData, error) {
	shield, err := shieldID()
	if err != nil {
		return nil, err
	}
	var pk packet.CraftingData
	if err := decodePacket(&pk, craftingData, shield); err != nil {
		return nil, fmt.Errorf("data: crafting data: %w", err)
	}
	return &pk, nil
}

// shieldID returns the shield's network ID, which item stacks on the wire
// depend on.
func shieldID() (int32, error) {
	items, err := ItemRegistry()
	if err != nil {
		return 0, err
	}
	for _, it := range items {
		if it.Name == "minecraft:shield" {
			return int32(it.RuntimeID), nil
		}
	}
	return 0, errors.New("data: no minecraft:shield in the item registry")
}

// decodePacket decodes an embedded packet payload into pk.
func decodePacket(pk packet.Packet, b []byte, shield int32) (err error) {
	defer func() {
		// The protocol reader stops on bad input by panicking; the data is
		// embedded and checked when written, so this only guards changes.
		if v := recover(); v != nil {
			err = fmt.Errorf("decode: %v", v)
		}
	}()
	buf := bytes.NewBuffer(b)
	pk.Marshal(protocol.NewReader(buf, shield, false))
	if buf.Len() != 0 {
		return fmt.Errorf("%d bytes left", buf.Len())
	}
	return nil
}
