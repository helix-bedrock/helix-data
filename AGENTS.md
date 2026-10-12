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
go run ./cmd/blockstates convert -world <bds>/worlds/helixdump -fill block_palette.nbt -out data/<version>
go run ./cmd/blockstates compare -world <bds>/worlds/helixdump -palette block_palette.nbt
```

## Layout

- `bds-dump/`: behaviour pack (Script API) that makes BDS save every block state into structures.
- `blockstate/`: reads those structures from the world, writes `block_states.nbt`, reads and compares palettes.
- `cmd/blockstates/`: command line for the above.
- `data/<version>/`: generated output. Never edit by hand.

## Invariants: do not break these

1. Every file in `data/` is generated from BDS by a tool in this repo, and
   `CREDITS.md` says which BDS version and which tool.
2. `block_states.nbt` uses pmmp's `canonical_block_states.nbt` layout:
   little-endian NBT compounds `{name, states, version}` one after another.
   Boolean properties are bytes, numbers are int32.
3. States come from BDS. The only exception is `-fill` for the blocks in
   `blockstate.DefaultFillBlocks`; never add to that list without showing
   BDS cannot produce the block, and credit the source.

## Gotchas

- You can run BDS yourself if the person points you at a BDS folder; work on
  a copy so their worlds stay untouched. You still cannot join the game.
- `BlockPermutation.getAllStates()` returns legacy alias properties
  (`stone_type` on `minecraft:stone`) and `BlockStates` value lists are shared
  between blocks. Only what BDS saves in a structure is trusted.
- Placing blocks in the world makes BDS recompute neighbour properties (fence
  connections, stair corners); the add-on writes into empty structures
  instead (`structureManager.createEmpty`, `@minecraft/server` 2.10.0).
- The server must be stopped with `stop`, or the structures never reach disk.
