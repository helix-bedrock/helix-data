package data

import (
	_ "embed"
	"encoding/json"
	"fmt"
)

//go:embed 1.26.60/bds/climbable.json
var climbable []byte

//go:embed 1.26.60/bds/block_materials.json
var blockMaterials []byte

// MaterialWater is BDS's MaterialType of water.
const MaterialWater = 5

// Climbable returns the block types BDS lets mobs and players climb
// (ladders, vines). Each call decodes a fresh copy.
func Climbable() ([]string, error) {
	var doc struct {
		Climbable []string `json:"climbable"`
	}
	if err := json.Unmarshal(climbable, &doc); err != nil {
		return nil, fmt.Errorf("data: climbable: %w", err)
	}
	return doc.Climbable, nil
}

// BlockMaterials returns each block type's BDS MaterialType (0 air, 5
// water, 6 lava...), keyed by block name. Each call decodes a fresh copy.
func BlockMaterials() (map[string]uint8, error) {
	var doc struct {
		Blocks map[string]uint8 `json:"blocks"`
	}
	if err := json.Unmarshal(blockMaterials, &doc); err != nil {
		return nil, fmt.Errorf("data: block materials: %w", err)
	}
	return doc.Blocks, nil
}

//go:embed 1.26.60/bds/placement.json
var placement []byte

// PlacementTrait is one of BDS's placement traits on a block: Kind
// "cardinal" or "facing" (the direction from the player's yaw, turned by
// Offset degrees), "vertical_half" or "block_face" (from the clicked
// face and position).
type PlacementTrait struct {
	Kind   string  `json:"kind"`
	Offset float32 `json:"offset"`
}

// Placement is how BDS picks a block's state when a player places it: the
// rule its own code follows ("stairs", "pillar", "torch"...; "pending"
// when the rule needs the world and is not ported) and the traits it
// applies.
type Placement struct {
	Rule   string           `json:"rule"`
	Traits []PlacementTrait `json:"traits"`
}

// PlacementData returns the placement of every block that is not simply
// placed in its item's state, keyed by block name. Each call decodes a
// fresh copy.
func PlacementData() (map[string]Placement, error) {
	var doc struct {
		Blocks map[string]Placement `json:"blocks"`
	}
	if err := json.Unmarshal(placement, &doc); err != nil {
		return nil, fmt.Errorf("data: placement: %w", err)
	}
	return doc.Blocks, nil
}

//go:embed 1.26.60/bds/piston_move.json
var pistonMove []byte

// PistonMove is what BDS's pistons do with blocks. A block in no list moves
// normally.
type PistonMove struct {
	// Immovable stops a piston, Breaks is destroyed by being pushed, NoPull is
	// pushed but never pulled, and Sticky takes its neighbours along.
	Immovable []string `json:"immovable"`
	Breaks    []string `json:"breaks"`
	NoPull    []string `json:"no_pull"`
	Sticky    []string `json:"sticky"`
}

// PistonMoveData returns the piston rules of [Version]. Each call decodes a
// fresh copy.
func PistonMoveData() (*PistonMove, error) {
	var doc PistonMove
	if err := json.Unmarshal(pistonMove, &doc); err != nil {
		return nil, fmt.Errorf("data: piston move: %w", err)
	}
	return &doc, nil
}

//go:embed 1.26.60/bds/support.json
var support []byte

// Support is what BDS lets blocks be attached to. Rules is each block's rule
// for which faces hold something attached (a block not listed holds nothing),
// HangFrom the blocks a lantern may also hang from, Solid the blocks a ladder
// may lean on.
type Support struct {
	Rules    map[string]string `json:"rules"`
	HangFrom []string          `json:"hang_from"`
	Solid    []string          `json:"solid"`
}

// SupportData returns BDS's attachment rules.
func SupportData() (*Support, error) {
	var doc Support
	if err := json.Unmarshal(support, &doc); err != nil {
		return nil, fmt.Errorf("data: support: %w", err)
	}
	return &doc, nil
}

//go:embed 1.26.60/bds/connections.json
var connections []byte

// Connection is a block's rule for what connects to it: Groups is a bit set of
// the connecting block's groups, Faces of the sides that accept it (1 << face,
// in the order down, up, north, south, west, east). States, when set, lists the
// rule of each state of a block whose rule depends on its state; a state with
// Groups -1 has no rule. A block with no rule lets everything connect.
type Connection struct {
	Groups int               `json:"groups"`
	Faces  int               `json:"faces"`
	States []StateConnection `json:"states,omitempty"`
}

// StateConnection is one state's Connection.
type StateConnection struct {
	States map[string]any `json:"states"`
	Groups int            `json:"groups"`
	Faces  int            `json:"faces"`
}

// ConnectionData returns the connection rule of every block with one, keyed by
// block name. Each call decodes a fresh copy.
func ConnectionData() (map[string]Connection, error) {
	var doc struct {
		Blocks map[string]Connection `json:"blocks"`
	}
	if err := json.Unmarshal(connections, &doc); err != nil {
		return nil, fmt.Errorf("data: connections: %w", err)
	}
	return doc.Blocks, nil
}

//go:embed 1.26.60/bds/liquids.json
var liquids []byte

// LiquidFacts are how BDS's liquids treat each block: Reaction is what a
// liquid flowing into the block does ("blocking", "broken": replaced,
// "popped": replaced and the block drops, "no_reaction": the liquid goes
// in the block's second layer), CanContain lists the blocks that can hold
// a liquid in their second layer, Solid the blocks whose material is
// solid, and FlowByShape the blocks whose own shape decides where a liquid
// in them may flow.
type LiquidFacts struct {
	Reaction    map[string][]string `json:"reaction"`
	CanContain  []string            `json:"can_contain"`
	Solid       []string            `json:"solid"`
	FlowByShape []string            `json:"flow_by_shape"`
	// LavaBlocking lists the blocks that stop lava flowing into them.
	LavaBlocking []string `json:"lava_blocking"`
}

// Liquids returns the liquid facts of every block. Each call decodes a
// fresh copy.
func Liquids() (LiquidFacts, error) {
	var doc LiquidFacts
	if err := json.Unmarshal(liquids, &doc); err != nil {
		return doc, fmt.Errorf("data: liquids: %w", err)
	}
	return doc, nil
}

//go:embed 1.26.60/bds/chest_obstruction.json
var chestObstruction []byte

// ChestObstruction lists the blocks that do not stop a chest under them opening
// whatever their shape: Blocks in every state, States in only those states. Any
// other block stops it if its collision boxes reach the lid.
type ChestObstruction struct {
	Blocks []string                    `json:"blocks"`
	States map[string][]map[string]any `json:"states"`
}

// ChestObstructionData returns BDS's chest obstruction lists. Each call decodes a
// fresh copy.
func ChestObstructionData() (ChestObstruction, error) {
	var doc ChestObstruction
	if err := json.Unmarshal(chestObstruction, &doc); err != nil {
		return ChestObstruction{}, fmt.Errorf("data: chest obstruction: %w", err)
	}
	return doc, nil
}

//go:embed 1.26.60/bds/block_drops.json
var blockDrops []byte

// DropNode is one step of a block's drop rule as BDS 1.26.60 builds it: an item
// list, a count or a test, by Kind. Which fields are set depends on Kind.
type DropNode struct {
	Kind    string    `json:"kind"`
	Name    string    `json:"name,omitempty"`
	State   string    `json:"state,omitempty"`
	States  []string  `json:"states,omitempty"`
	Value   int32     `json:"value,omitempty"`
	Add     int32     `json:"add,omitempty"`
	N       int32     `json:"n,omitempty"`
	Min     int32     `json:"min,omitempty"`
	Max     int32     `json:"max,omitempty"`
	Table   []float64 `json:"table,omitempty"`
	Seeds   string    `json:"seeds,omitempty"`
	Product string    `json:"product,omitempty"`
	Count   *DropNode `json:"count,omitempty"`
	Of      *DropNode `json:"of,omitempty"`
	A       *DropNode `json:"a,omitempty"`
	B       *DropNode `json:"b,omitempty"`
	If      *DropNode `json:"if,omitempty"`
	Then    *DropNode `json:"then,omitempty"`
	Else    *DropNode `json:"else,omitempty"`
}

// BlockDrop is what a block drops: Drops is its rule, or nil if the block drops
// by its loot table in vanilla/loot_tables/blocks. XP is its experience range.
type BlockDrop struct {
	XP    [2]int32  `json:"xp"`
	Drops *DropNode `json:"drops,omitempty"`
}

// BlockDrops returns BDS's drop rule of every block, by block name. Each call
// decodes a fresh copy.
func BlockDrops() (map[string]BlockDrop, error) {
	var doc struct {
		Blocks map[string]BlockDrop `json:"blocks"`
	}
	if err := json.Unmarshal(blockDrops, &doc); err != nil {
		return nil, fmt.Errorf("data: block drops: %w", err)
	}
	return doc.Blocks, nil
}

//go:embed 1.26.60/bds/redstone_conductivity.json
var redstoneConductivity []byte

// BlockConductivity is what one block does with redstone power: Power is
// "conductor", "insulator" or empty. A block whose answer depends on its state
// names that State, with the answer per state value in Values.
type BlockConductivity struct {
	Power  string            `json:"power,omitempty"`
	State  string            `json:"state,omitempty"`
	Values map[string]string `json:"values,omitempty"`
}

// RedstoneConductivity returns BDS's redstone conductivity by block name. Blocks
// not in it take no part.
func RedstoneConductivity() (map[string]BlockConductivity, error) {
	var d struct {
		Blocks map[string]BlockConductivity `json:"blocks"`
	}
	if err := json.Unmarshal(redstoneConductivity, &d); err != nil {
		return nil, fmt.Errorf("data: read redstone conductivity: %w", err)
	}
	return d.Blocks, nil
}

//go:embed 1.26.60/bds/note_instruments.json
var noteInstruments []byte

// NoteInstruments is which instrument a note block plays. Above holds the blocks that
// play on top of it, Below the instrument of the block under it, 0 if not listed.
type NoteInstruments struct {
	Above map[string]uint8 `json:"above"`
	Below map[string]uint8 `json:"below"`
}

// NoteInstrumentData returns the note block instruments of [Version]. Each call decodes
// a fresh copy.
func NoteInstrumentData() (*NoteInstruments, error) {
	var doc NoteInstruments
	if err := json.Unmarshal(noteInstruments, &doc); err != nil {
		return nil, fmt.Errorf("data: note instruments: %w", err)
	}
	return &doc, nil
}

//go:embed 1.26.60/bds/music_discs.json
var musicDiscs []byte

// MusicDisc is what a disc does in a jukebox.
type MusicDisc struct {
	// Sound names the song. The client plays it from the jukebox's block entity.
	Sound string `json:"sound"`
	// Seconds is how long the jukebox counts as playing.
	Seconds float32 `json:"seconds"`
	// Signal is what a comparator reads from a jukebox holding the disc.
	Signal int32 `json:"signal"`
}

// MusicDiscData returns the music discs of [Version] by item name. Each call decodes a
// fresh copy.
func MusicDiscData() (map[string]MusicDisc, error) {
	var doc map[string]MusicDisc
	if err := json.Unmarshal(musicDiscs, &doc); err != nil {
		return nil, fmt.Errorf("data: music discs: %w", err)
	}
	return doc, nil
}

//go:embed 1.26.60/bds/block_push.json
var blockPush []byte

// BlockPush is what BDS's push of items and orbs out of blocks reads of a block type:
// Solid lists those that push and that count as closed sides, Stairs those that push
// more slowly.
type BlockPush struct {
	Solid  []string `json:"solid"`
	Stairs []string `json:"stairs"`
}

// BlockPushData returns BDS's block facts for pushing out of blocks.
func BlockPushData() (*BlockPush, error) {
	var doc BlockPush
	if err := json.Unmarshal(blockPush, &doc); err != nil {
		return nil, fmt.Errorf("data: block push: %w", err)
	}
	return &doc, nil
}

//go:embed 1.26.60/bds/furnace.json
var furnaceFacts []byte

// Furnace is what BDS's furnaces, blast furnaces and smokers read of items. Fuel is
// how many cooks one item burns for, by item name: a furnace burns it for its cook
// time times that (coal blocks 80, logs and planks 1.5). FuelSeconds are the items
// whose fuel is a minecraft:fuel component instead, burning for the cook time times
// seconds over ten. XP is the experience one smelted item gives, by the item made.
// CookTicks is each block's time to smelt one item.
type Furnace struct {
	CookTicks   map[string]int32   `json:"cook_ticks"`
	Fuel        map[string]float32 `json:"fuel"`
	FuelSeconds map[string]float32 `json:"fuel_seconds"`
	XP          map[string]float32 `json:"xp"`
}

// FurnaceData returns BDS's furnace facts. Each call decodes a fresh copy.
func FurnaceData() (*Furnace, error) {
	var doc Furnace
	if err := json.Unmarshal(furnaceFacts, &doc); err != nil {
		return nil, fmt.Errorf("data: furnace: %w", err)
	}
	return &doc, nil
}

//go:embed 1.26.60/bds/banner.json
var bannerFacts []byte

// BannerPattern is one of BDS's banner patterns: its name, the ID banners save and
// looms are asked for, and the type of banner pattern item a loom needs for it (-1:
// none, the pattern is drawn with a dye alone).
type BannerPattern struct {
	Name     string `json:"name"`
	ID       string `json:"id"`
	ItemType int16  `json:"item_type"`
}

// Banner is what BDS's banners and looms read: the patterns in BDS's order, how many
// one banner holds, the dye items with their colour (dye order, black 0 to white 15)
// and the banner pattern items with their type.
type Banner struct {
	MaxPatterns  int              `json:"max_patterns"`
	Patterns     []BannerPattern  `json:"patterns"`
	Dyes         map[string]uint8 `json:"dyes"`
	PatternItems map[string]int16 `json:"pattern_items"`
}

// BannerData returns BDS's banner facts. Each call decodes a fresh copy.
func BannerData() (*Banner, error) {
	var doc Banner
	if err := json.Unmarshal(bannerFacts, &doc); err != nil {
		return nil, fmt.Errorf("data: banner: %w", err)
	}
	return &doc, nil
}

//go:embed 1.26.60/bds/precipitation.json
var precipitation []byte

// Precipitation is BDS's precipitation behaviour of a block: what rain and snow do at it.
type Precipitation uint8

// The precipitation behaviours. A block BDS gives none is PrecipitationAccumulateSnow.
const (
	// PrecipitationAccumulateSnow stops rain and snow settles on top (most blocks).
	PrecipitationAccumulateSnow Precipitation = iota
	// PrecipitationNone lets rain and snow fall through (air, torches, crops).
	PrecipitationNone
	// PrecipitationObstructRain stops rain, but snow never settles on it (ice, dirt paths).
	PrecipitationObstructRain
	// PrecipitationSnowlogging lets rain through, and snow settles into it as a covered snow
	// layer (flowers, short grass).
	PrecipitationSnowlogging
)

// BlockPrecipitation returns each block type's precipitation behaviour, keyed by block name;
// block types not in the map are PrecipitationAccumulateSnow. Each call decodes a fresh copy.
func BlockPrecipitation() (map[string]Precipitation, error) {
	var doc struct {
		None         []string `json:"none"`
		ObstructRain []string `json:"obstruct_rain"`
		Snowlogging  []string `json:"snowlogging"`
	}
	if err := json.Unmarshal(precipitation, &doc); err != nil {
		return nil, fmt.Errorf("data: precipitation: %w", err)
	}
	m := make(map[string]Precipitation, len(doc.None)+len(doc.ObstructRain)+len(doc.Snowlogging))
	for _, n := range doc.None {
		m[n] = PrecipitationNone
	}
	for _, n := range doc.ObstructRain {
		m[n] = PrecipitationObstructRain
	}
	for _, n := range doc.Snowlogging {
		m[n] = PrecipitationSnowlogging
	}
	return m, nil
}
