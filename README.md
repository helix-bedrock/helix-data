# helix-data

Minecraft: Bedrock Edition game data for Helix, and the tools that build it
from the official Bedrock Dedicated Server (BDS).

**Status as of 2026-10-25:** in development. `data/1.26.52/vanilla/` holds every vanilla behaviour pack file Helix reads, stacked as BDS 1.26.52.3 stacks them, and `data/1.26.52/bds/` the world generator tables read from BDS's code.

Part of [Helix](https://github.com/helix-bedrock), an open-source Minecraft: Bedrock
Edition server engine in Go.

## What is here

| Path | What it is |
|---|---|
| `bds-dump/` | A BDS behaviour pack that prints every block state to the server console |
| `cmd/blockstates` | Turns that console log into `block_states.nbt` and `block_defaults.json`, and compares it with other palettes |
| `blockstate` | The Go package behind `cmd/blockstates` |
| `data/<version>/` | Generated files, one folder per Minecraft version |

## Dumping block states from BDS

1. Download BDS for the target version from minecraft.net and run it once.
2. Copy `bds-dump/` to `behavior_packs/helix-dump/` in the BDS folder.
3. In `server.properties`, set `level-name=helixdump` and
   `content-log-console-output-enabled=true`.
4. Create `worlds/helixdump/world_behavior_packs.json`:
   ```json
   [{ "pack_id": "6f2b8f0e-4c1a-4b7e-9a51-1d3c5e7a9b20", "version": [1, 0, 0] }]
   ```
5. Start BDS. When the console prints `HELIXDUMP-END` (a few seconds), type
   `stop` so the world is written to disk.
6. Convert it:
   ```bash
   go run ./cmd/blockstates convert -world <bds>/worlds/helixdump -out data/<version>
   go run ./cmd/blockstates compare -world <bds>/worlds/helixdump -palette block_palette.nbt
   ```
7. For items, save the BDS console output while the dump runs (for example
   `bedrock_server > out.log`) and convert it:
   ```bash
   go run ./cmd/blockstates items -log out.log -out data/<version>
   ```
8. Entities, their loot tables and spawn rules need no dump: they are read straight from
   the BDS folder.
   ```bash
   go run ./cmd/entities -bds <bds folder> -out data/<version>
   ```
9. The vanilla behaviour pack files (biomes, features, feature rules, all
   loot tables, recipes, spawn groups, structures, trading and jigsaw
   world generation) need no dump either. BDS ships one pack per game
   version holding only what changed; `cmd/vanilla` stacks them the way BDS
   does and writes one complete `data/<version>/vanilla/`:
   ```bash
   go run ./cmd/vanilla -bds <bds folder> -out data/<version>
   ```
   Go code reads them, and the BDS tables in `data/<version>/bds/`, through
   `data.Files()`.

### How it works

The add-on writes every combination of every block's properties into
structures and saves them to the world. BDS stores each block in a structure
palette exactly as it would on disk, with its version number, and rejects
combinations a block does not have. `cmd/blockstates` reads those palettes
with `helix-bedrock/world`'s LevelDB.

`minecraft:structure_void` cannot be stored in a structure, so the add-on
places one in the world and the converter reads it from the chunk on disk.

### Compared with CloudburstMC/Data (1.26.52)

`compare` against CloudburstMC/Data's 1.26.50 `block_palette.nbt`: 22,058
states are identical. The differences, all checked against BDS:

- Cloudburst lists `minecraft:fire`, `redstone_wire` and `chorus_plant` with
  connection properties (1,872 states). BDS 1.26.52 never stores or sends
  those properties: built next to real neighbours, these blocks were saved
  and sent to a joined client without them. The 33 forms BDS does use are in
  our data and not in Cloudburst's.
- `disc_press`, `ice_crystal` and `icicle` (30 states) do not exist in BDS
  1.26.52, with or without experiments.

`blockstate.NetworkHash` computes the network ID of a state; its test pins
IDs BDS 1.26.52.3 sent to a client.

## Contributing

See the [contributing guide](https://github.com/helix-bedrock/.github/blob/main/CONTRIBUTING.md).
AI-assisted contributions follow the
[AI policy](https://github.com/helix-bedrock/.github/blob/main/AI_POLICY.md).

## License

The code is MIT; see [`LICENSE`](LICENSE). Generated data describes Minecraft
and comes from Mojang's server; where each file comes from is listed in
[`CREDITS.md`](CREDITS.md).

## World generation data

BDS ships no world generation data, so `worldgen.json` and
`biome_tree.json` come from Java Edition, whose terrain and biomes match
Bedrock's since 1.18 (see CREDITS.md):

```bash
go run ./cmd/worldgen -datapack <datapack>/data/minecraft/worldgen -tree <multi_noise_biome_tree.json> -out data/<version>
```
