import csv
import json
import os
from pathlib import Path
import shutil
import signal
import subprocess
import sys
import time

ROOT = Path('/private/tmp/tsgolint-node-sort')
OUT = Path('/Users/cameron/.codex/worktrees/b7ca/tsgolint/benchmarks/checker-affinity/node-sort-results/profiles')
REPO = Path('/Users/cameron/.codex/worktrees/b7ca/tsgolint')
EVENTS = OUT / 'profile-load-events.json'
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
        if command.startswith(str(ROOT)) and name in ('stealing.test', 'sorted.test'):
            continue
        other_cpu += float(cpu)
        is_builder = name in BUILDERS
        if name == 'go':
            # A pprof HTTP server is a long-lived viewer, not a build.
            # Its CPU still contributes to other_cpu above.
            command_line = subprocess.run(['ps', '-p', pid, '-o', 'args='],
                                          text=True, capture_output=True).stdout.strip()
            if command_line.startswith('go tool pprof ') and '-http=' in command_line:
                is_builder = False
        if is_builder:
            builders.append({'pid': int(pid), 'name': name})
    (OUT / 'profile-load-status.json').write_text(json.dumps({'time': time.time(), 'builders': builders, 'other_cpu': round(other_cpu, 1)}))
    return builders, round(other_cpu, 1)

def ready():
    quiet_since = None
    announced = False
    while True:
        builders, cpu = load()
        if builders or cpu > 400:
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

event('monitor_policy', sample_interval_seconds=5, quiet_seconds=30, sustained_busy_seconds=10, other_cpu_threshold=400, ignore_pprof_http_as_builder=True, phase='profile')

first_attempt = 1 + max((e.get('attempt', 0) for e in events), default=0)
for attempt in range(first_attempt, first_attempt + 100):
    ready()
    command = [sys.executable, str(OUT / 'run-profiles.py')]
    with (ROOT / f'node-sort-profiles-attempt-{attempt}.log').open('w') as log:
        process = subprocess.Popen(command, cwd=REPO, stdout=log, stderr=subprocess.STDOUT, start_new_session=True)
        event('started', attempt=attempt, pid=process.pid)
        busy_since = None
        interrupted = False
        while process.poll() is None:
            before = time.time()
            builders, cpu = load()
            if builders or cpu > 400:
                if busy_since is None:
                    busy_since = before - 5
                if before - busy_since >= 10:
                    event('contention', attempt=attempt, builders=builders, other_cpu=cpu)
                    try:
                        os.killpg(process.pid, signal.SIGTERM)
                    except ProcessLookupError:
                        pass
                    process.wait(timeout=10)
                    event('profile_interrupted', attempt=attempt)
                    interrupted = True
                    break
            else:
                busy_since = None
            time.sleep(5)
        source = OUT / 'environment-profile.json'
        if source.exists():
            shutil.copyfile(source, OUT / f'environment-profiles-guard-{attempt}.json')
        if not interrupted:
            if process.returncode != 0:
                event('failed', attempt=attempt, exit_code=process.returncode)
                raise SystemExit(process.returncode)
            event('completed', attempt=attempt)
            break
else:
    raise SystemExit('too many contention interruptions')
