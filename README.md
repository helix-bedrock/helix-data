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
3. In `worlds/<level-name>/`, create `world_behavior_packs.json`:
   ```json
   [{ "pack_id": "6f2b8f0e-4c1a-4b7e-9a51-1d3c5e7a9b20", "version": [1, 0, 0] }]
   ```
   Use a throwaway world; nobody needs to join.
4. Start BDS with its output saved to a file, for example
   `bedrock_server.exe > dump.log` (Windows) or `./bedrock_server > dump.log` (Linux).
5. Wait until the log contains `HELIXDUMP-END`, then type `stop`.
6. Convert it:
   ```bash
   go run ./cmd/blockstates convert -log dump.log -version 18168865 -out data/1.26.52
   go run ./cmd/blockstates compare -dump dump.log -palette block_palette.nbt
   ```

`-version` is the block version number stored in each state. The Script API
cannot read it, so it is passed in; for 1.26.50+ it is `18168865`, the value
in CloudburstMC/Data's 1.26.50 palette.

## Contributing

See the [contributing guide](https://github.com/helix-bedrock/.github/blob/main/CONTRIBUTING.md).
AI-assisted contributions follow the
[AI policy](https://github.com/helix-bedrock/.github/blob/main/AI_POLICY.md).

## License

The code is MIT; see [`LICENSE`](LICENSE). Generated data describes Minecraft
and comes from Mojang's server; where each file comes from is listed in
[`CREDITS.md`](CREDITS.md).
