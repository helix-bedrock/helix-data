# Credits

Sources of the code and data in this repo, as required by the
[porting rules](https://github.com/helix-bedrock/.github/blob/main/PORTING.md).

## Code

No ported code. `bds-dump`, `brarchive`, `cmd/blockstates` and `cmd/entities` were written for Helix.

## Data

| Path | Source | How |
|---|---|---|
| `data/1.26.52/block_states.nbt`, `block_defaults.json` | Bedrock Dedicated Server 1.26.52.3 (Mojang) | Saved by BDS through `bds-dump`, converted by `cmd/blockstates convert` |
| `data/1.26.52/items.json` | Bedrock Dedicated Server 1.26.52.3 (Mojang) | Printed by BDS through `bds-dump` (`ItemTypes.getAll()`, `ItemStack.maxAmount`), converted by `cmd/blockstates items` |
| `data/1.26.52/entities.json`, `entity_loot.json`, `spawn_rules.json` | Bedrock Dedicated Server 1.26.52.3 (Mojang) | Read from BDS's vanilla behaviour packs (`__brarchive/entities.brarchive`, `spawn_rules.brarchive`, `loot_tables/entities`) by `cmd/entities`, packs layered as BDS layers them |

[CloudburstMC/Data](https://github.com/CloudburstMC/Data)'s 1.26.50
`block_palette.nbt` (commit `a8a4341`) is used only to compare against; none
of it is copied into this repo. See the README for the comparison.
