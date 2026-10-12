# Credits

Sources of the code and data in this repo, as required by the
[porting rules](https://github.com/helix-bedrock/.github/blob/main/PORTING.md).

## Code

No ported code. `bds-dump` and `cmd/blockstates` were written for Helix.

## Data

| Path | Source | How |
|---|---|---|
| `data/<version>/block_states.nbt`, `block_defaults.json` | Bedrock Dedicated Server (Mojang) of that version | Printed by `bds-dump` through the Script API, converted by `cmd/blockstates` |

Comparisons use [CloudburstMC/Data](https://github.com/CloudburstMC/Data)'s
`block_palette.nbt`, which is not copied into this repo.
