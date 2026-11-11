# AGENTS.md

Rules for AI coding agents working in this repo. Also read the org-wide rules in
`../.github/AGENT_RULES.md` (online:
<https://github.com/helix-bedrock/.github/blob/main/AGENT_RULES.md>). Where they
disagree, this file wins for this repo.

## What this repo is

Game data for Helix generated from BDS, and the tools that generate it. It
holds data and small tools only; no engine code. `helix-bedrock/world` and the
engine read the files in `data/`.

## Commands

```bash
go build ./...
go vet ./...
golangci-lint run
go test -race ./...
go run ./cmd/blockstates convert -world <bds>/worlds/helixdump -out data/files
go run ./cmd/blockstates compare -world <bds>/worlds/helixdump -palette block_palette.nbt
go run ./cmd/blockstates items -log <bds console log> -out data/files
go run ./cmd/entities -bds <bds folder> -out data/files
go run ./cmd/vanilla -bds <bds folder> -out data/files
go run ./cmd/joinpackets -capture <join.jsonl> -out data/files/bds/packets
go run ./cmd/worldgen -datapack <datapack>/data/minecraft/worldgen -tree <multi_noise_biome_tree.json> -out data/files
```

## Layout

- `bds-dump/`: behaviour pack (Script API) that makes BDS save every block state into structures and print every item type with its stack size.
- `item/`: reads the item lines of the BDS console log and writes and reads `items.json`.
- `blockstate/`: reads those structures (and structure_void from the chunk) from the world, writes `block_states.nbt`, reads and compares palettes, and computes network hashes.
- `brarchive/`: reads the `.brarchive` files BDS packs behaviour pack folders into.
- `pack/`: stacks BDS's vanilla behaviour packs (`vanilla`, then `vanilla_<version>` oldest first, over `definitions/`) the way BDS does, unpacking their archives.
- `cmd/vanilla/`: writes `vanilla/` from the stacked packs: biomes, features, feature_rules, loot_tables, recipes, spawn_groups, structures, trading and worldgen, with `_provenance.tsv` naming each file's pack, and `experiments/<name>/` with each `experimental_*` pack's own files, unstacked (BDS loads an enabled experiment's pack above every vanilla pack).
- `cmd/entities/`: writes `entities.json`, `entity_loot.json` (loot_tables/entities and loot_tables/gameplay), `spawn_rules.json` and `item_definitions.json` (data-driven item definitions, such as bundles' storage components) from BDS's vanilla behaviour packs.
- `mcb/`: reads the binary (starting with the bytes 0x7F "MCB") form some item definitions take in `.brarchive` files into their JSON form. Components are named by the FNV-1a hash of their name, with fields in a fixed order; an unknown component fails the run, so add its fields to `components` when BDS adds one.
- `cmd/worldgen/`: writes `worldgen.json` and `biome_tree.json` from Java Edition's vanilla datapack.
- `cmd/joinpackets/`: writes `bds/packets/` from a capture of a real join on BDS: the raw payloads of the registry packets, each checked to decode and encode back byte for byte.
- `cmd/blockstates/`: command line for the above, plus `experiments` to turn on experiment toggles in a world's level.dat.
- `data/`: the `data` Go package, which embeds the current version's files so the engine can load them.
- `data/files/`: generated output. Never edit by hand.

## Invariants: do not break these

1. Every file in `data/` comes from BDS, made by a tool in this repo or,
   for `bds/`, read out of BDS with the workspace's dump skills. The one
   exception is world generation (`worldgen.json`, `biome_tree.json`),
   which comes from Java Edition and is credited in `CREDITS.md`. Data
   taken from BDS gets no `CREDITS.md` entry (decided by the maintainer,
   2026-11-01).
2. `block_states.nbt` uses pmmp's `canonical_block_states.nbt` layout:
   little-endian NBT compounds `{name, states, version}` one after another.
   Boolean properties are bytes, numbers are int32.
3. Every state comes from BDS. Other palettes are only compared against,
   never copied in.
4. `data/files/` is complete on its own. BDS ships only the
   files that changed in each pack; `cmd/vanilla` stacks them so nothing
   that reads the data has to know about the older packs.

## Gotchas

- You can run BDS yourself if the person points you at a BDS folder; work on
  a copy so their worlds stay untouched. You still cannot join the game.
- `BlockPermutation.getAllStates()` returns legacy alias properties
  (`stone_type` on `minecraft:stone`) and `BlockStates` value lists are shared
  between blocks. Only what BDS saves in a structure is trusted.
- Placing blocks in the world makes BDS recompute neighbour properties (fence
  connections, stair corners); the add-on writes into empty structures
  instead (`structureManager.createEmpty`, `@minecraft/server` 2.10.0).
- Other palettes listed fire, redstone_wire and chorus_plant connection
  properties that BDS 1.26.52 never stores or sends. BDS 1.26.60 has them, so
  `data/files/` holds them.
- From 1.26.60 the add-on misses a few states (awake `creaking_heart`,
  `sculk_sensor_phase` 2) because saving a structure changes them. Read
  block states from the running server (`update-block-states` skill) and use
  the add-on only to check them.
- The server must be stopped with `stop`, or the structures never reach disk.
- `.brarchive` files list binary structures (`.nbt`, `.mcstructure`) as
  empty entries; the real file sits next to the archive. `pack` skips empty
  entries and lets plain files win within a pack.
- BDS 1.26.52.3 builds desert villages with the base pack's three
  `desert_meeting_point_*.nbt`, not `vanilla_1.20.0`'s. `cmd/vanilla` pins
  them (`pinned`); add any other such file there with how it was measured.
