# Frozen experiment sources

`build.py` uses these files as Go overlays. The `.txt` extension keeps snapshots
out of normal Go package discovery. All variants use the current committed
`cmd/tsgolint/checker_affinity_bench_test.go` harness.

- `baseline-linter.go.txt`: original shared queue from TSGolint
  `acd88e1b1ef8f219e8956f804f2548dafd949705`.
- `affinity-linter.go.txt`: owner queues with no stealing, from the pure-affinity POC.
- `stealing-linter.go.txt`: owner queues with stealing, matching this branch's linter.
- `unsorted-checkerpool.go.txt`: compiler after patches 0001–0005, before node sorting.
- `sorted-checkerpool.go.txt`: compiler after patch 0006 (descending node count).

The first four snapshots match their counterparts in `node-sort-results/`, and
all five match the scheduling-source hashes in `node-sort-results/source.json`.
The historical harness/runner snapshots there preserve that measurement's inputs;
the portable runner uses the current harness, which also excludes imported JSON.
Keep these snapshots fixed when reproducing the experiment; to test a new
algorithm, make an explicit additional variant and use a new results directory.
