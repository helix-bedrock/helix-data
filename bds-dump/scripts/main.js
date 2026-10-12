// Prints every block state BDS knows to the server console, one JSON object
// per line, for cmd/blockstates in helix-data to turn into data files.
//
// Lines:
//   HELIXDUMP-BEGIN {"blocks":N}
//   HELIXDUMP-DEFAULT {"name":"minecraft:stone","states":{...}}   one per block
//   HELIXDUMP {"name":"minecraft:stone","states":{...}}           one per state
//   HELIXDUMP-SKIP {"name":...,"states":{...},"error":"..."}       combination BDS rejected
//   HELIXDUMP-END {"blocks":N,"states":N,"skipped":N}
//
// The work runs as a job spread over many ticks so the script watchdog does
// not stop it.
import { BlockPermutation, BlockStates, BlockTypes, system, world } from "@minecraft/server";

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

function* dump() {
  const types = BlockTypes.getAll().map((t) => t.id).sort();
  console.log("HELIXDUMP-BEGIN " + JSON.stringify({ blocks: types.length }));
  let states = 0;
  let skipped = 0;
  for (const id of types) {
    let def;
    try {
      def = BlockPermutation.resolve(id);
    } catch (e) {
      console.log("HELIXDUMP-SKIP " + JSON.stringify({ name: id, states: {}, error: String(e) }));
      skipped++;
      continue;
    }
    const defaults = def.getAllStates();
    console.log("HELIXDUMP-DEFAULT " + JSON.stringify({ name: id, states: defaults }));

    // Only the states the default permutation has belong to this block;
    // their valid values come from the global state registry.
    const names = Object.keys(defaults).sort();
    const values = names.map((n) => {
      const t = BlockStates.get(n);
      return t ? t.validValues : [defaults[n]];
    });
    for (const combo of combinations(names, values, 0, {})) {
      try {
        const p = BlockPermutation.resolve(id, combo);
        console.log("HELIXDUMP " + JSON.stringify({ name: id, states: p.getAllStates() }));
        states++;
      } catch (e) {
        console.log("HELIXDUMP-SKIP " + JSON.stringify({ name: id, states: combo, error: String(e) }));
        skipped++;
      }
      yield;
    }
  }
  console.log("HELIXDUMP-END " + JSON.stringify({ blocks: types.length, states, skipped }));
}

world.afterEvents.worldLoad.subscribe(() => {
  system.runJob(dump());
});
