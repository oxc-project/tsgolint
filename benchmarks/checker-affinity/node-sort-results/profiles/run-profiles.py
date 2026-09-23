"""Untimed full-project profiles isolating the node-count assignment change."""

import hashlib
import json
from pathlib import Path
from types import SimpleNamespace
import sys

EXPERIMENT = Path(__file__).resolve().parents[2]
sys.path.insert(0, str(EXPERIMENT))
from run import Experiment  # noqa: E402

ROOT = Path('/private/tmp/tsgolint-node-sort')
OUT = Path(__file__).resolve().parent
args = SimpleNamespace(
    output=OUT, variant=None, baseline=ROOT / 'stealing.test',
    candidate=ROOT / 'sorted.test',
    projects=Path('/private/tmp/tsgolint-affinity-w4vcz79f/projects'),
    names=['typescript', 'vscode', 'typeorm', 'vue'],
)
experiment = Experiment(args)
manifests = experiment.manifests()
metadata = {
    'labels': {'baseline': 'affinity + stealing', 'candidate': 'affinity + stealing + node sort'},
    'workers': 4, 'gomaxprocs': 4, 'semantic': False, 'selection': 'full',
    'binaries': {v: hashlib.sha256(Path(p).read_bytes()).hexdigest()
                 for v, p in experiment.binaries.items()},
}
(OUT / 'environment-profile.json').write_text(json.dumps(metadata, indent=2) + '\n')
results = []
for name in args.names:
    case = experiment.case(name)
    case.update(files=manifests[name]['full'], mode='profile')
    for version in experiment.binaries:
        profiles = experiment.invoke(version, case)
        total = sum(p['rule_ns'] for p in profiles)
        results.append({
            'project': name, 'selection': 'full', 'semantic': False,
            'version': version, 'checkers': profiles,
            'active_checkers': sum(p['files'] > 0 for p in profiles),
            'slowest_share': max(p['rule_ns'] for p in profiles) / total if total else 0,
        })
        (OUT / 'profiles.json').write_text(json.dumps(results, indent=2) + '\n')
    print('profiled', name, flush=True)
