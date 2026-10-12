# Credits

Sources of the code and data in this repo, as required by the
[porting rules](https://github.com/helix-bedrock/.github/blob/main/PORTING.md).

## Code

No ported code. `bds-dump` and `cmd/blockstates` were written for Helix.

## Data

| Path | Source | How |
|---|---|---|
| `data/1.26.52/block_states.nbt`, `block_defaults.json` | Bedrock Dedicated Server 1.26.52.3 (Mojang) | Saved by BDS through `bds-dump`, converted by `cmd/blockstates convert` |

[CloudburstMC/Data](https://github.com/CloudburstMC/Data)'s 1.26.50
`block_palette.nbt` (commit `a8a4341`) is used only to compare against; none
of it is copied into this repo. See the README for the comparison.
