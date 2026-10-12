# helix-data

Minecraft: Bedrock Edition game data for Helix, and the tools that build it
from the official Bedrock Dedicated Server (BDS).

**Status:** in development

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
   go run ./cmd/blockstates convert -world <bds>/worlds/helixdump -fill block_palette.nbt -out data/<version>
   go run ./cmd/blockstates compare -world <bds>/worlds/helixdump -palette block_palette.nbt
   ```

### How it works

The add-on writes every combination of every block's properties into
structures and saves them to the world. BDS stores each block in a structure
palette exactly as it would on disk, with its version number, and rejects
combinations a block does not have. `cmd/blockstates` reads those palettes
with `helix-bedrock/world`'s LevelDB.

The Script API cannot set a few properties, so `-fill` adds the states of
these blocks from CloudburstMC/Data's palette (`blockstate.DefaultFillBlocks`):

- `minecraft:fire`, `minecraft:redstone_wire`, `minecraft:chorus_plant`: their
  connection properties are computed from neighbours. BDS (and the game) save
  them on disk without those properties; the forms with them come from
  Cloudburst. Both forms are kept.
- `minecraft:structure_void`: cannot be stored in a structure.

Experimental blocks (`disc_press`, `ice_crystal`, `icicle` in 1.26.52) are not
included.

## Contributing

See the [contributing guide](https://github.com/helix-bedrock/.github/blob/main/CONTRIBUTING.md).
AI-assisted contributions follow the
[AI policy](https://github.com/helix-bedrock/.github/blob/main/AI_POLICY.md).

## License

The code is MIT; see [`LICENSE`](LICENSE). Generated data describes Minecraft
and comes from Mojang's server; where each file comes from is listed in
[`CREDITS.md`](CREDITS.md).
