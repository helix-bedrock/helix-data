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
go run ./cmd/blockstates convert -log dump.log -version 18168865 -out data/<version>
go run ./cmd/blockstates compare -dump dump.log -palette block_palette.nbt
```

## Layout

- `bds-dump/`: behaviour pack (Script API) that prints block states to the BDS console.
- `blockstate/`: parses the dump, writes `block_states.nbt`, reads and compares palettes.
- `cmd/blockstates/`: command line for the above.
- `data/<version>/`: generated output. Never edit by hand.

## Invariants: do not break these

1. Every file in `data/` is generated from BDS by a tool in this repo, and
   `CREDITS.md` says which BDS version and which tool.
2. `block_states.nbt` uses pmmp's `canonical_block_states.nbt` layout:
   little-endian NBT compounds `{name, states, version}` one after another.
   Boolean properties are bytes, numbers are int32.
3. A dump without a `HELIXDUMP-END` line is rejected, never half-converted.

## Gotchas

- You cannot run BDS. A human runs the dump and gives you the log; say which
  steps you did not run.
- The Script API cannot read block version numbers or network runtime IDs;
  `-version` is passed in, and runtime IDs are computed by the engine.
- `bds-dump` lists every combination of a block's properties and asks BDS to
  resolve it; combinations BDS rejects are logged as `HELIXDUMP-SKIP`.
  Check those after each new version.
