// Makes BDS save every block state it knows into the world database, for
// cmd/blockstates in helix-data to read.
//
// The Script API cannot list block states directly: getAllStates() also
// returns legacy alias properties (stone_type on minecraft:stone) and
// resolve() accepts values a block does not use. So every combination is
// written into a structure and saved; BDS writes the real state,
// with its version number, into the structure's palette. Duplicates and
// aliases disappear there.
//
// structure_void cannot be stored in a structure, so one is left in the
// world for the converter to read from the chunk on disk.
//
// Structures are saved to the world as helixdump:p<N> (all combinations) and
// helixdump:d<N> (default permutations only). Console lines:
//   HELIXDUMP-BEGIN {"blocks":N}
//   HELIXDUMP-BATCH {"id":"helixdump:p1","placed":N}
//   HELIXDUMP-SKIP {"name":...,"states":{...},"error":"..."}
//   HELIXDUMP-ITEM {"name":...,"maxStack":N}
//   HELIXDUMP-END {"blocks":N,"placed":N,"structures":N,"skipped":N,"items":N}
// After HELIXDUMP-END, stop the server with "stop" so the world is written.
import { BlockPermutation, BlockStates, BlockTypes, ItemStack, ItemTypes, StructureSaveMode, system, world } from "@minecraft/server";

// Each structure holds one SIDE x 1 x SIDE layer.
const SIDE = 64;
const PER_BATCH = SIDE * SIDE;
const MAX_EXTRA_INT = 15;

// structure_void is placed here, in a ticking area so the chunk is loaded.
const VOID_POS = { x: 0, y: 250, z: 0 };

function* waitLoaded(dim, pos) {
  for (;;) {
    try {
      if (dim.getBlock(pos) !== undefined) return;
    } catch {
      // Not loaded yet.
    }
    yield;
  }
}

function* combinations(names, values, i, current) {
  if (i === names.length) {
    yield { ...current };
    return;
  }
  for (const v of values[i]) {
    current[names[i]] = v;
    yield* combinations(names, values, i + 1, current);
  }
  delete current[names[i]];
}

function* permutations(types, onSkip) {
  for (const id of types) {
    let def;
    try {
      def = BlockPermutation.resolve(id);
    } catch (e) {
      onSkip(id, {}, e);
      continue;
    }
    const defaults = def.getAllStates();
    const names = Object.keys(defaults).sort();
    const values = names.map((n) => {
      const t = BlockStates.get(n);
      const v = t ? Array.from(t.validValues ?? []) : [];
      if (v.length && typeof v[0] === "number") {
        // validValues is shared by every block with this property, but some
        // blocks allow more (minecraft:chalkboard's direction goes to 15).
        // Try 0..15 too; BDS rejects values a block does not allow.
        for (let i = 0; i <= MAX_EXTRA_INT; i++) if (!v.includes(i)) v.push(i);
      }
      return v.length ? v : [defaults[n]];
    });
    for (const combo of combinations(names, values, 0, {})) {
      try {
        yield BlockPermutation.resolve(id, combo);
      } catch (e) {
        onSkip(id, combo, e);
      }
    }
  }
}

function* defaults(types) {
  for (const id of types) {
    try {
      yield BlockPermutation.resolve(id);
    } catch {
      // Reported by permutations().
    }
  }
}

// Writes up to PER_BATCH permutations straight into a new structure and
// saves it to the world. Nothing is placed in the world, so BDS never
// recomputes properties that depend on neighbours (fence connections, stair
// corners). Returns how many were written.
function saveBatch(id, batch, onSkip) {
  const s = world.structureManager.createEmpty(id, { x: SIDE, y: 1, z: SIDE }, StructureSaveMode.Memory);
  let placed = 0;
  for (let i = 0; i < batch.length; i++) {
    try {
      s.setBlockPermutation({ x: i % SIDE, y: 0, z: Math.floor(i / SIDE) }, batch[i]);
      placed++;
    } catch (e) {
      onSkip(batch[i].type.id, batch[i].getAllStates(), e);
    }
  }
  s.saveToWorld();
  console.log("HELIXDUMP-BATCH " + JSON.stringify({ id, placed }));
  return placed;
}

function* dump() {
  const types = BlockTypes.getAll().map((t) => t.id).sort();
  console.log("HELIXDUMP-BEGIN " + JSON.stringify({ blocks: types.length }));
  let skipped = 0;
  const onSkip = (name, states, e) => {
    skipped++;
    console.log("HELIXDUMP-SKIP " + JSON.stringify({ name, states, error: String(e) }));
  };

  let placed = 0;
  let structures = 0;
  for (const [prefix, source] of [["p", permutations(types, onSkip)], ["d", defaults(types)]]) {
    let batch = [];
    let n = 0;
    for (const p of source) {
      batch.push(p);
      if (batch.length === PER_BATCH) {
        placed += saveBatch(`helixdump:${prefix}${++n}`, batch, onSkip);
        structures++;
        batch = [];
        yield;
      }
    }
    if (batch.length) {
      placed += saveBatch(`helixdump:${prefix}${++n}`, batch, onSkip);
      structures++;
    }
    yield;
  }
  const dim = world.getDimension("overworld");
  dim.runCommand(`tickingarea add ${VOID_POS.x} ${VOID_POS.y} ${VOID_POS.z} ${VOID_POS.x} ${VOID_POS.y} ${VOID_POS.z} helixvoid true`);
  yield* waitLoaded(dim, VOID_POS);
  dim.getBlock(VOID_POS).setPermutation(BlockPermutation.resolve("minecraft:structure_void"));

  // Items: every item type and how many fit in one stack. The Script API
  // reports BDS's own value, so no stack size is guessed.
  const items = ItemTypes.getAll().map((t) => t.id).sort();
  for (const name of items) {
    try {
      const maxStack = new ItemStack(name, 1).maxAmount;
      console.log("HELIXDUMP-ITEM " + JSON.stringify({ name, maxStack }));
    } catch (e) {
      console.log("HELIXDUMP-SKIP " + JSON.stringify({ name, item: true, error: String(e) }));
    }
  }

  console.log("HELIXDUMP-END " + JSON.stringify({ blocks: types.length, placed, structures, skipped, items: items.length }));
}

world.afterEvents.worldLoad.subscribe(() => {
  system.runJob(dump());
});
