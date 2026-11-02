# Credits

Sources of the code and data in this repo, as required by the
[porting rules](https://github.com/helix-bedrock/.github/blob/main/PORTING.md).

## Code

No ported code. `bds-dump`, `brarchive`, `pack`, `cmd/blockstates`, `cmd/entities`, `cmd/vanilla` and `cmd/worldgen` were written for Helix.

## Data

### World generation (Java Edition)

These two files come from Java Edition, whose terrain shape and biomes match
Bedrock's for the same seed since 1.18:

| Path | Source | How |
|---|---|---|
| `data/1.26.52/worldgen.json`, `data/1.26.60/worldgen.json` | Java Edition 26.3 server jar, vanilla datapack `data/minecraft/worldgen/{density_function,noise,noise_settings,material_rule,material_condition}` | Bundled by `cmd/worldgen`. The datapack was extracted by Pumpkin's build (github.com/Pumpkin-MC/Pumpkin, `assets/datapack`) |
| `data/1.26.52/biome_tree.json`, `data/1.26.60/biome_tree.json` | Java Edition 26.3, the overworld and nether biome parameter search trees, which Java builds in code | Extracted from Java Edition by Pumpkin's extractor (Pumpkin `assets/multi_noise_biome_tree.json`), compacted by `cmd/worldgen` |

Bedrock's own world generation has not been compared against these yet.

[CloudburstMC/Data](https://github.com/CloudburstMC/Data)'s 1.26.50
`block_palette.nbt` (commit `a8a4341`) is used only to compare against; none
of it is copied into this repo. See the README for the comparison.
