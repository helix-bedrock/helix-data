# Credits

Sources of the code and data in this repo, as required by the
[porting rules](https://github.com/helix-bedrock/.github/blob/main/PORTING.md).

## Code

No ported code. `bds-dump`, `brarchive`, `pack`, `cmd/blockstates`, `cmd/entities`, `cmd/vanilla` and `cmd/worldgen` were written for Helix.

## Data

| Path | Source | How |
|---|---|---|
| `data/1.26.52/block_states.nbt`, `block_defaults.json` | Bedrock Dedicated Server 1.26.52.3 (Mojang) | Saved by BDS through `bds-dump`, converted by `cmd/blockstates convert` |
| `data/1.26.52/items.json` | Bedrock Dedicated Server 1.26.52.3 (Mojang) | Printed by BDS through `bds-dump` (`ItemTypes.getAll()`, `ItemStack.maxAmount`), converted by `cmd/blockstates items` |
| `data/1.26.52/entities.json`, `entity_loot.json`, `spawn_rules.json` | Bedrock Dedicated Server 1.26.52.3 (Mojang) | Read from BDS's vanilla behaviour packs (`__brarchive/entities.brarchive`, `spawn_rules.brarchive`, `loot_tables/entities`) by `cmd/entities`, packs layered as BDS layers them |

| `data/1.26.52/vanilla/` | Bedrock Dedicated Server 1.26.52.3 (Mojang) | The vanilla behaviour packs and `definitions/` stacked as BDS stacks them, by `cmd/vanilla`; `_provenance.tsv` names each file's pack |
| `data/1.26.60/block_states.nbt`, `block_defaults.json` | Bedrock Dedicated Server 1.26.60.6 Linux (Mojang) | Read from the running server under gdb (every Block's serialization tag when BDS hashes it for the network), converted by the workspace's `update-block-states` skill; checked against `bds-dump` |
| `data/1.26.60/items.json` | Bedrock Dedicated Server 1.26.60.6 (Mojang) | Printed by BDS through `bds-dump`, converted by `cmd/blockstates items` |
| `data/1.26.60/entities.json`, `entity_loot.json`, `spawn_rules.json` | Bedrock Dedicated Server 1.26.60.6 (Mojang) | Read from BDS's vanilla behaviour packs by `cmd/entities` |
| `data/1.26.60/vanilla/` | Bedrock Dedicated Server 1.26.60.6 (Mojang) | The vanilla behaviour packs and `definitions/` stacked by `cmd/vanilla` |
| `data/1.26.60/bds/` | Bedrock Dedicated Server 1.26.52.3 Linux (Mojang) | Copied from `data/1.26.52/bds/` until the tables are read from 1.26.60.6; each file's `source` names the build |
| `data/1.26.52/bds/` | Bedrock Dedicated Server 1.26.52.3 Linux (Mojang) | Tables read out of BDS's code with IDA for the world generator (jigsaw pools and elements, block rules, liquid holders, top support), and `block_materials.json`, read from the running server under a debugger; each file's `source` names the BDS functions. Moved from `helix-bedrock/worldgen` |

### World generation (Java Edition)

BDS ships no world generation data, so these two files come from Java
Edition (Mojang), whose terrain shape and biomes match Bedrock's for the same
seed since 1.18:

| Path | Source | How |
|---|---|---|
| `data/1.26.52/worldgen.json`, `data/1.26.60/worldgen.json` | Java Edition 26.3 server jar, vanilla datapack `data/minecraft/worldgen/{density_function,noise,noise_settings,material_rule,material_condition}` | Bundled by `cmd/worldgen`. The datapack was extracted by Pumpkin's build (github.com/Pumpkin-MC/Pumpkin, `assets/datapack`) |
| `data/1.26.52/biome_tree.json`, `data/1.26.60/biome_tree.json` | Java Edition 26.3, the overworld and nether biome parameter search trees, which Java builds in code | Extracted from Java Edition by Pumpkin's extractor (Pumpkin `assets/multi_noise_biome_tree.json`), compacted by `cmd/worldgen` |

Bedrock's own world generation has not been compared against these yet.

[CloudburstMC/Data](https://github.com/CloudburstMC/Data)'s 1.26.50
`block_palette.nbt` (commit `a8a4341`) is used only to compare against; none
of it is copied into this repo. See the README for the comparison.
