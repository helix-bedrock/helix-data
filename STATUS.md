# Status

Status as of 2026-11-12: in development.

The `data` package serves `data/files/` (BDS 1.26.60.6; only the newest version is kept):

- 23,960 block states of 1,480 blocks, read from the running server.
- 1,623 items.
- Every vanilla behaviour pack file Helix reads, stacked as BDS stacks them, plus each
  experimental pack's files under `experiments/` (Drop 4 of 2026's ice caves among them).

## BDS tables read from 1.26.60.6

| File | What it holds |
|---|---|
| `bds/block_properties.json` | Every state's light, flammability, explosion resistance, friction and destroy time, read from the running server |
| `bds/block_types.json` | Each block's translucency (BDS keeps it on the block type) |
| `bds/map_colors.json` | Each block's map colour as BDS's map render reads it: the map_color component's colour and tint method, or the block type's colour; log top and side colours; beds from their block entity |
| `bds/block_solid.json` | The blocks whose Block solid flag is set (full opaque blocks, which a respawn anchor's spawn search stands players on) |
| `bds/precipitation.json` | Each block's precipitation behaviour: what rain and snow do at it (fall through, stop, settle as snow, or snowlog) |
| `bds/compost.json` | The composter's item chances (BDS's built-in table, read from the running server) |
| `bds/biome_ids.json` | Every vanilla biome's numeric ID as BDS registers it |
| `bds/item_equipment.json` | Which items BDS lets a player wear in each armour slot, and which go in the offhand |
| `bds/mining.json` | What BDS's mining speed rules read: every block type's tags and whether it needs the right tool, each tool's tag and tier, the sword and shears block lists |
| `bds/break_place.json` | Each item's max damage, how it wears when breaking blocks, and the blocks a placed block may replace |
| `bds/collision.json` | The collision boxes of the 851 blocks whose shape is fixed (stored as data in BDS, or a constant in its code), and for the 629 whose shape BDS computes in code, the rule each follows |
| `bds/packets/` | The registry packets BDS sends every joining player: 2,080 items with their network IDs, creative inventory, crafting, biomes, entity identifiers, the 98 data-driven block definitions from StartGame, and the voxel shapes and jigsaw structure data sent before it |

The other `bds/` tables are still the 1.26.52.3 ones (each file's `source` says so) until they
are read again from 1.26.60.6.
