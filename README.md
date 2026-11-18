# helix-data

Minecraft: Bedrock Edition game data for Helix, and the tools that build it
from the official Bedrock Dedicated Server (BDS).

**Status:** in development. What `data/files/` holds now, file by file: [STATUS.md](STATUS.md).

Part of [Helix](https://github.com/helix-bedrock), a source-available Minecraft: Bedrock
Edition server engine in Go.

## What is here

| Path | What it is |
|---|---|
| `bds-dump/` | A BDS behaviour pack that prints every block state to the server console |
| `cmd/blockstates` | Turns that console log into `block_states.nbt` and `block_defaults.json`, and compares it with other palettes |
| `blockstate` | The Go package behind `cmd/blockstates` |
| `data/files/` | Generated files of the newest Minecraft version (older versions are not kept) |

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
   go run ./cmd/blockstates convert -world <bds>/worlds/helixdump -out data/files
   go run ./cmd/blockstates compare -world <bds>/worlds/helixdump -palette block_palette.nbt
   ```
7. For items, save the BDS console output while the dump runs (for example
   `bedrock_server > out.log`) and convert it:
   ```bash
   go run ./cmd/blockstates items -log out.log -out data/files
   ```
8. Entities, their loot tables and spawn rules need no dump: they are read straight from
   the BDS folder.
   ```bash
   go run ./cmd/entities -bds <bds folder> -out data/files
   ```
9. The vanilla behaviour pack files (biomes, features, feature rules, all
   loot tables, recipes, spawn groups, structures, trading and jigsaw
   world generation) need no dump either. BDS ships one pack per game
   version holding only what changed; `cmd/vanilla` stacks them the way BDS
   does and writes one complete `data/files/vanilla/`:
   ```bash
   go run ./cmd/vanilla -bds <bds folder> -out data/files
   ```
   It also writes `data/files/experiments/<name>/`: the files of each
   experimental pack (`experimental_y_2026_drop_4` becomes `y_2026_drop_4`)
   on their own. BDS loads an enabled experiment's pack above every vanilla
   pack, so a reader puts these files over `vanilla/`.
   Go code reads them, and the BDS tables in `data/files/bds/`, through
   `data.Files()`.

### How it works

The add-on writes every combination of every block's properties into
structures and saves them to the world. BDS stores each block in a structure
palette exactly as it would on disk, with its version number, and rejects
combinations a block does not have. `cmd/blockstates` reads those palettes
with `helix-bedrock/world`'s LevelDB.

`minecraft:structure_void` cannot be stored in a structure, so the add-on
places one in the world and the converter reads it from the chunk on disk.

### 1.26.60

BDS 1.26.60.6 adds `disc_press`, `ice_crystal` and `icicle`, and the
connection properties of `fire`, `redstone_wire` and `chorus_plant` (1,902
states added, 33 removed). The states were read from the running server
(the `update-block-states` skill) and checked against the add-on: the same
1,480 blocks and defaults. The add-on misses 7 states (awake
`creaking_heart`, `sculk_sensor_phase` 2) that BDS turns into other states
when saving a structure; 1.26.52's data has them too.

`blockstate.NetworkHash` computes the network ID of a state; its test pins
IDs BDS 1.26.52.3 sent to a client.

## Contributing

See the [contributing guide](https://github.com/helix-bedrock/.github/blob/main/CONTRIBUTING.md).
AI-assisted contributions follow the
[AI policy](https://github.com/helix-bedrock/.github/blob/main/AI_POLICY.md).

## License

Noncommercial use only: [PolyForm Noncommercial 1.0.0](LICENSE) with Helix's
[extra permissions](ADDITIONAL-PERMISSIONS.md). Anyone may use it, run servers with it
(even ones that make money) and sell their own plugins, but nobody may sell Helix or its
code. Helix's game behaviour is re-engineered from Minecraft's Bedrock Dedicated Server;
see [NOTICE.md](NOTICE.md). Not an official Minecraft product. Not approved by or
associated with Mojang or Microsoft.

The game data files belong to Mojang and are not covered by this license. Sources are
listed in [`CREDITS.md`](CREDITS.md).

