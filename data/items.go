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

//go:embed 1.26.60/bds/packets/start_game_blocks.bin
var startGameBlocks []byte

// BlockDefinitions returns the block definitions BDS sends joining players
// in StartGame: the vanilla blocks defined by data (components, states,
// geometry) rather than built into the client. Without them the client
// does not know those blocks.
func BlockDefinitions() ([]protocol.BlockEntry, error) {
	var pk blockList
	if err := decodePacket(&pk, startGameBlocks, 0); err != nil {
		return nil, fmt.Errorf("data: start game blocks: %w", err)
	}
	return pk.blocks, nil
}

// blockList reads StartGame's block list on its own.
type blockList struct{ blocks []protocol.BlockEntry }

func (*blockList) ID() uint32               { return 0 }
func (b *blockList) Marshal(io protocol.IO) { protocol.Slice(io, &b.blocks) }

//go:embed 1.26.60/bds/packets/voxel_shapes.bin
var voxelShapes []byte

// VoxelShapes returns the VoxelShapes packet BDS sends joining players
// before StartGame: the named block shapes the client uses.
func VoxelShapes() (*packet.VoxelShapes, error) {
	var pk packet.VoxelShapes
	if err := decodePacket(&pk, voxelShapes, 0); err != nil {
		return nil, fmt.Errorf("data: voxel shapes: %w", err)
	}
	return &pk, nil
}

//go:embed 1.26.60/bds/packets/jigsaw_structure_data.bin
var jigsawStructureData []byte

// JigsawStructureData returns the JigsawStructureData packet BDS sends
// joining players before StartGame.
func JigsawStructureData() (*packet.JigsawStructureData, error) {
	var pk packet.JigsawStructureData
	if err := decodePacket(&pk, jigsawStructureData, 0); err != nil {
		return nil, fmt.Errorf("data: jigsaw structure data: %w", err)
	}
	return &pk, nil
}

//go:embed 1.26.60/bds/item_equipment.json
var itemEquipment []byte

// ItemEquipment returns, as BDS checks them, the armour slot of every item
// a player may wear (0 head, 1 chest, 2 legs, 3 feet) and the items the
// offhand accepts. Each call decodes a fresh copy.
func ItemEquipment() (armorSlots map[string]uint8, offhand []string, err error) {
	var doc struct {
		ArmorSlots map[string]uint8 `json:"armorSlots"`
		Offhand    []string         `json:"offhand"`
	}
	if err := json.Unmarshal(itemEquipment, &doc); err != nil {
		return nil, nil, fmt.Errorf("data: item equipment: %w", err)
	}
	return doc.ArmorSlots, doc.Offhand, nil
}

//go:embed 1.26.60/bds/mining.json
var mining []byte

// Mining is what BDS's mining speed rules read: the items whose speed BDS
// works out in code and every block type's tags.
type Mining struct {
	// TierTags are the block tags that ask for a digger of at least tier
	// 1, 2, 3 and 4 to count as the right tool.
	TierTags []string `json:"tierTags"`
	// Items are the tools, keyed by item name.
	Items map[string]MiningItem `json:"items"`
	// Sword, Shears and Mace are the block lists and tags of those rules.
	Sword  MiningSword  `json:"sword"`
	Shears MiningShears `json:"shears"`
	Mace   struct {
		Tag string `json:"tag"`
	} `json:"mace"`
	// Blocks are every block type's tags and whether it needs the right
	// tool, keyed by block name.
	Blocks map[string]MiningBlock `json:"blocks"`
}

// MiningItem is a tool: Rule is "digger", "sword", "shears" or "mace". A
// digger mines blocks with Tag at Speed and counts as the right tool for
// blocks whose tier tag asks for at most Tier.
type MiningItem struct {
	Rule  string  `json:"rule"`
	Tag   string  `json:"tag"`
	Tier  int     `json:"tier"`
	Speed float32 `json:"speed"`
}

// MiningSword lists the blocks swords mine at 15, 30 and 1.5; Tag marks
// the blocks a sword is the right tool for.
type MiningSword struct {
	Tag      string   `json:"tag"`
	Speed15  []string `json:"speed15"`
	Speed30  []string `json:"speed30"`
	Speed1_5 []string `json:"speed1_5"`
}

// MiningShears lists the blocks shears mine at 15, 5 and 2 plus the
// Efficiency bonus (Tag also gives 5); Tag and CanDestroy mark the blocks
// shears are the right tool for.
type MiningShears struct {
	Tag        string   `json:"tag"`
	CanDestroy []string `json:"canDestroy"`
	Plus15     []string `json:"plus15"`
	Plus5      []string `json:"plus5"`
	Plus2      []string `json:"plus2"`
}

// MiningBlock is a block type's tags and whether it needs the right tool.
type MiningBlock struct {
	Tags      []string `json:"tags"`
	NeedsTool bool     `json:"needsTool"`
}

// MiningData returns the mining table. Each call decodes a fresh copy.
func MiningData() (*Mining, error) {
	var m Mining
	if err := json.Unmarshal(mining, &m); err != nil {
		return nil, fmt.Errorf("data: mining: %w", err)
	}
	return &m, nil
}
