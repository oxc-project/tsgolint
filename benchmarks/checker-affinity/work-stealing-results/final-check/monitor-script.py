import csv
import json
import os
from pathlib import Path
import shutil
import signal
import subprocess
import sys
import time

ROOT = Path('/private/tmp/tsgolint-affinity-w4vcz79f')
OUT = ROOT / 'results-steal-final-check'
REPO = Path('/Users/cameron/.codex/worktrees/b7ca/tsgolint')
EVENTS = OUT / 'control-load-events.json'
OUT.mkdir(exist_ok=True)
events = json.loads(EVENTS.read_text()) if EVENTS.exists() else []
BUILDERS = {'go', 'cargo', 'rustc', 'compile', 'link', 'clang', 'clang++'}

def event(kind, **data):
    row = {'time': time.time(), 'kind': kind, **data}
    events.append(row)
    EVENTS.write_text(json.dumps(events, indent=2) + '\n')
    print(json.dumps(row), flush=True)

def load():
    raw = subprocess.check_output(['ps', '-A', '-o', 'pid=,ppid=,%cpu=,comm='], text=True)
    builders = []
    other_cpu = 0
    for line in raw.splitlines():
        fields = line.strip().split(None, 3)
        if len(fields) != 4:
            continue
        pid, ppid, cpu, command = fields
        name = Path(command).name
        if command.startswith(str(ROOT)) and name in ('steal-v3.test', 'steal-final.test'):
            continue
        other_cpu += float(cpu)
        if name in BUILDERS:
            builders.append({'pid': int(pid), 'name': name})
    (OUT / 'control-load-status.json').write_text(json.dumps({'time': time.time(), 'builders': builders, 'other_cpu': round(other_cpu, 1)}))
    return builders, round(other_cpu, 1)

def ready():
    quiet_since = None
    announced = False
    while True:
        builders, cpu = load()
        if builders or cpu > 200:
            quiet_since = None
            if not announced:
                event('waiting', builders=builders, other_cpu=cpu)
                announced = True
        else:
            if quiet_since is None:
                quiet_since = time.monotonic()
            if time.monotonic() - quiet_since >= 30:
                return
        time.sleep(5)

def discard_affected(cutoff, attempt):
    path = OUT / 'samples.csv'
    shutil.copyfile(path, ROOT / f'quarantined-steal-final-check-attempt-{attempt}.csv')
    with path.open() as stream:
        reader = csv.DictReader(stream)
        fields = reader.fieldnames
        rows = list(reader)
    groups = {}
    for row in rows:
        key = (row['project'], row['scenario'], row['worker_setting'])
        groups.setdefault(key, []).append(row)
    excluded = [r for r in rows if float(r.get('started_at') or 0) + float(r.get('elapsed_seconds') or 0) >= cutoff]
    kept = [r for r in rows if r not in excluded]
    excluded_cases = sorted({(r['project'], r['scenario'], r['worker_setting']) for r in excluded})
    temporary = path.with_suffix('.guard-tmp')
    with temporary.open('w', newline='') as stream:
        writer = csv.DictWriter(stream, fieldnames=fields)
        writer.writeheader()
        writer.writerows(kept)
    temporary.replace(path)
    event('excluded', cases=excluded_cases, samples=len(rows) - len(kept), retained_ordered_prefix=True)

event('monitor_policy', sample_interval_seconds=5, quiet_seconds=30, sustained_busy_seconds=10, other_cpu_threshold=200, retain_uncontended_prefix=True)

first_attempt = 1 + max((e.get('attempt', 0) for e in events), default=0)
for attempt in range(first_attempt, first_attempt + 100):
    ready()
    command = [sys.executable, 'benchmarks/checker-affinity/run.py',
               '--baseline', str(ROOT / 'steal-v3.test'), '--candidate', str(ROOT / 'steal-final.test'),
               '--projects', str(ROOT / 'projects'), '--output', str(OUT),
               '--phase', 'measure', '--workers', '4', '--names', 'typescript', '--scenarios', 'full-lint', 'full-check', '--samples', '6', '--resume-partial']
    with (ROOT / f'steal-final-check-attempt-{attempt}.log').open('w') as log:
        process = subprocess.Popen(command, cwd=REPO, stdout=log, stderr=subprocess.STDOUT, start_new_session=True)
        event('started', attempt=attempt, pid=process.pid)
        busy_since = None
        interrupted = False
        while process.poll() is None:
            before = time.time()
            builders, cpu = load()
            if builders or cpu > 200:
                if busy_since is None:
                    busy_since = before - 5
                if before - busy_since >= 10:
                    event('contention', attempt=attempt, builders=builders, other_cpu=cpu)
                    try:
                        os.killpg(process.pid, signal.SIGTERM)
                    except ProcessLookupError:
                        pass
                    process.wait(timeout=10)
                    discard_affected(busy_since, attempt)
                    interrupted = True
                    break
            else:
                busy_since = None
            time.sleep(5)
        source = OUT / 'environment-measure.json'
        if source.exists():
            shutil.copyfile(source, OUT / f'environment-controls-guard-{attempt}.json')
        if not interrupted:
            if process.returncode != 0:
                event('failed', attempt=attempt, exit_code=process.returncode)
                raise SystemExit(process.returncode)
            event('completed', attempt=attempt)
            break
else:
    raise SystemExit('too many contention interruptions')
