# Credits

Sources of the code and data in this repo, as required by the
[porting rules](https://github.com/helix-bedrock/.github/blob/main/PORTING.md).

## Code

No ported code. `bds-dump` and `cmd/blockstates` were written for Helix.

## Data

| Path | Source | How |
|---|---|---|
| `data/1.26.52/block_states.nbt`, `block_defaults.json` | Bedrock Dedicated Server 1.26.52.3 (Mojang) | Saved by BDS through `bds-dump`, converted by `cmd/blockstates convert` |
| The `minecraft:fire`, `minecraft:redstone_wire`, `minecraft:chorus_plant` and `minecraft:structure_void` states in `data/1.26.52/block_states.nbt` that BDS did not save | [CloudburstMC/Data](https://github.com/CloudburstMC/Data) `block_palette.nbt` at [`a8a4341`](https://github.com/CloudburstMC/Data/commit/a8a4341) (1.26.50) | Added by `cmd/blockstates convert -fill`; see the README for why |

Comparisons also use that Cloudburst palette, which is not copied into this
repo otherwise.

### 1.26.52 check

`compare` against Cloudburst's 1.26.50 palette: 22,057 states identical; the
only differences are the four fill blocks above and the experimental blocks
`disc_press`, `ice_crystal` and `icicle`, which Cloudburst lists and BDS does
not enable by default.
