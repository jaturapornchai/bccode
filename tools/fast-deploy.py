#!/usr/bin/env python3
"""
Fast Zero-Disk Streamed Production Deployment for BC Account.
Optimized for speed:
- Streaming `docker save` directly through compressed SSH pipe into remote `docker load` (no temporary tar files on disk)
- When backend is unchanged, re-tags remote mainapi image in 0ms (no rebuild or 300MB upload)
- Full preflight safety: live database backup (Mongo + Postgres) and atomic release.env switch
"""

import argparse
import datetime
import json
import os
import subprocess
import sys
import time
import urllib.error
import urllib.request

if hasattr(sys.stdout, "reconfigure"):
    sys.stdout.reconfigure(encoding="utf-8", errors="replace")
if hasattr(sys.stderr, "reconfigure"):
    sys.stderr.reconfigure(encoding="utf-8", errors="replace")

SERVER_HOST = "root@159.223.43.229"
SSH_BASE = ["ssh", "-o", "BatchMode=yes", "-o", "ConnectTimeout=15", SERVER_HOST]

def log(msg: str):
    now = datetime.datetime.now().strftime("%H:%M:%S")
    print(f"[{now}] {msg}", flush=True)

def run_cmd(args, cwd=None):
    return subprocess.check_output(args, cwd=cwd, text=True).strip()

def run_ssh(cmd_str: str) -> str:
    res = subprocess.run(SSH_BASE + [cmd_str], capture_output=True, text=True)
    if res.returncode != 0:
        raise RuntimeError(f"Remote command failed: {cmd_str}\nStderr: {res.stderr}\nStdout: {res.stdout}")
    return res.stdout.strip()

def main():
    parser = argparse.ArgumentParser(description="Fast Deploy BC AI Account")
    parser.add_argument("--tag", default="", help="Release tag (e.g. r20260912-gl-crud-1)")
    parser.add_argument("--all", action="store_true", help="Rebuild and deploy both backend and frontend")
    parser.add_argument("--frontend", action="store_true", help="Force frontend-only deploy (re-tag existing remote mainapi)")
    args = parser.parse_args()

    today_str = datetime.date.today().strftime("%Y%m%d")
    tag = args.tag.strip() or f"r{today_str}-gl-crud-1"

    workspace = os.path.abspath(os.path.join(os.path.dirname(__file__), ".."))
    frontend_dir = os.path.join(workspace, "frontend")
    backend_dir = os.path.join(workspace, "backend")

    start_total = time.time()
    log(f"🚀 Starting fast deploy for release: {tag}")

    # 1. Check if backend needs rebuild
    deploy_backend = args.all and not args.frontend
    if args.frontend:
        log("⚡ --frontend specified -> Frontend-only fast path forced (skipping mainapi rebuild/upload).")
    elif not deploy_backend:
        # Check git status for backend code changes
        try:
            diff_out = run_cmd(["git", "diff", "--name-only", "HEAD", "--", "backend"], cwd=workspace)
            untracked = run_cmd(["git", "status", "--porcelain", "--", "backend"], cwd=workspace)
            # Only code files in backend
            code_changes = [f for f in (diff_out + "\n" + untracked).splitlines() if f.endswith(".go") or f.endswith(".mod")]
            if code_changes:
                log(f"Backend code changes detected: {code_changes}. Will build both.")
                deploy_backend = True
            else:
                log("⚡ Backend code unchanged -> Frontend-only fast path enabled (skipping mainapi rebuild/upload).")
        except Exception as e:
            log(f"Git check notice: {e}, sticking to frontend-only default.")

    # 2. Check local vs remote Docker daemon readiness
    use_remote_docker = False
    try:
        check_local = subprocess.run(["docker", "info"], capture_output=True, timeout=4)
        if check_local.returncode != 0:
            use_remote_docker = True
    except Exception:
        use_remote_docker = True

    docker_base = ["docker"]
    if use_remote_docker:
        log("🌐 Local Docker Desktop unavailable -> Auto-switching to direct build via remote Docker daemon over SSH...")
        docker_base = ["docker", "--host", f"ssh://{SERVER_HOST}"]

    # 3. Build frontend image
    log("🔨 [Step 1/7] Building frontend docker image...")
    t0 = time.time()
    frontend_img = f"bcai-account-frontend:{tag}"
    build_frontend_cmd = docker_base + [
        "build",
        "--build-arg", "BCAI_LOCAL_BACKEND_URL=http://mainapi:8888",
        "--build-arg", "NEXT_PUBLIC_GOOGLE_CLIENT_ID=212036599086-c7aqvm005jiv2kqi4duju8spd9b3jb94.apps.googleusercontent.com",
        "-t", frontend_img,
        frontend_dir
    ]
    subprocess.check_call(build_frontend_cmd)
    log(f"✅ Frontend image built in {time.time() - t0:.1f}s")

    mainapi_img = f"bcai-account-mainapi:{tag}"
    if deploy_backend:
        log("🔨 Building backend docker image...")
        t0 = time.time()
        build_backend_cmd = docker_base + ["build", "-t", mainapi_img, backend_dir]
        subprocess.check_call(build_backend_cmd)
        log(f"✅ Backend image built in {time.time() - t0:.1f}s")

    # 4. Preflight backup on remote server
    log("🔒 [Step 2/7] Running remote preflight backup (Mongo + Postgres + Runtime Config)...")
    t0 = time.time()
    preflight_script = f"""
import datetime, hashlib, json, os, pathlib, subprocess
os.umask(0o077)
release = pathlib.Path('/opt/bcai-account/releases/{tag}')
release.mkdir(mode=0o700, parents=True, exist_ok=True)
backup = release / 'backup'
backup.mkdir(mode=0o700, exist_ok=True)

def run(args): return subprocess.check_output(args, text=True).strip()
def inspect(name): return json.loads(run(['docker', 'inspect', name]))[0]

main = inspect('bcai-account-mainapi-1')
frontend = inspect('bcai-account-frontend-1')
worker_img = ''
try:
    worker = inspect('bcai-account-worker-1')
    worker_img = worker['Config']['Image']
except Exception:
    pass
pg = inspect('bcai-account-postgres-1')
pg_env = dict(item.split('=', 1) for item in pg['Config']['Env'] if '=' in item)

def save_cmd(filename, cmd):
    target = backup / filename
    if target.exists(): target.unlink()
    with target.open('xb') as out:
        subprocess.run(cmd, stdout=out, check=True)
    with target.open('rb') as s:
        digest = hashlib.file_digest(s, 'sha256').hexdigest()
    return {{'filename': filename, 'bytes': target.stat().st_size, 'sha256': digest}}

artifacts = [
    save_cmd('postgres-all.sql', ['docker', 'exec', 'bcai-account-postgres-1', 'pg_dumpall', '-U', pg_env['POSTGRES_USER']]),
    save_cmd('runtime-config.tar.gz', ['tar', '-czf', '-', '/etc/bcai-account', '/var/lib/bcai-account/config', '/opt/bcai-account/deploy'])
]
try:
    subprocess.check_call(['docker', 'inspect', 'bcai-account-mongo-1'], stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
    artifacts.append(save_cmd('mongo.archive.gz', ['docker', 'exec', 'bcai-account-mongo-1', 'mongodump', '--archive', '--gzip', '--oplog']))
except Exception:
    pass

subprocess.run(['cp', '-p', '/etc/bcai-account/release.env', str(release / 'release.env.before')], check=True)
os.chmod(release / 'release.env.before', 0o600)

proof = {{
    'capturedat': datetime.datetime.now(datetime.timezone.utc).isoformat(),
    'oldimages': {{'mainapi': main['Config']['Image'], 'worker': worker_img, 'frontend': frontend['Config']['Image']}},
    'backups': artifacts
}}
(release / 'preflight-backup.json').write_text(json.dumps(proof, ensure_ascii=False, indent=2))
print(json.dumps(proof))
"""
    preflight_res = subprocess.run(
        SSH_BASE + ["python3", "-"],
        input=preflight_script,
        capture_output=True,
        text=True,
        check=True
    )
    proof = json.loads(preflight_res.stdout.strip().splitlines()[-1])
    current_mainapi_image = proof["oldimages"]["mainapi"]
    log(f"✅ Preflight backup complete in {time.time() - t0:.1f}s (Current mainapi: {current_mainapi_image})")

    # 5. Handle mainapi tag or upload
    if not deploy_backend:
        log(f"🏷️ [Step 3/7] Re-tagging remote mainapi to {mainapi_img} (Zero-copy instant tag)...")
        run_ssh(f"docker tag {current_mainapi_image} {mainapi_img}")
    else:
        if not use_remote_docker:
            log(f"📦 [Step 3/7] Streaming backend image {mainapi_img} over compressed SSH pipe...")
            t0 = time.time()
            p1 = subprocess.Popen(["docker", "save", mainapi_img], stdout=subprocess.PIPE)
            p2 = subprocess.Popen(SSH_BASE + ["-C", "docker", "load"], stdin=p1.stdout, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
            p1.stdout.close()
            out, err = p2.communicate()
            if p2.returncode != 0:
                raise RuntimeError(f"Failed to stream mainapi image: {err.decode()}")
            log(f"✅ Mainapi image loaded on remote server in {time.time() - t0:.1f}s: {out.decode().strip()}")
        else:
            log(f"⚡ [Step 3/7] Backend image {mainapi_img} built directly on remote daemon -> Zero transfer time!")

    # 6. Stream frontend image directly through compressed SSH pipe (if built locally)
    if not use_remote_docker:
        log(f"📦 [Step 4/7] Streaming frontend image {frontend_img} directly over compressed SSH pipe...")
        t0 = time.time()
        p1 = subprocess.Popen(["docker", "save", frontend_img], stdout=subprocess.PIPE)
        p2 = subprocess.Popen(SSH_BASE + ["-C", "docker", "load"], stdin=p1.stdout, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
        p1.stdout.close()
        out, err = p2.communicate()
        if p2.returncode != 0:
            raise RuntimeError(f"Failed to stream frontend image: {err.decode()}")
        log(f"✅ Frontend image loaded on remote server in {time.time() - t0:.1f}s: {out.decode().strip()}")
    else:
        log(f"⚡ [Step 4/7] Frontend image {frontend_img} built directly on remote daemon -> Zero transfer time!")

    # 7. Remote atomic switch and container restart
    log("🔄 [Step 5/7] Performing atomic release.env switch and container deployment...")
    t0 = time.time()
    deploy_script = f"""
import datetime, json, os, pathlib, re, subprocess, time, urllib.request, urllib.error
os.umask(0o077)
release = pathlib.Path('/opt/bcai-account/releases/{tag}')
release_env = pathlib.Path('/etc/bcai-account/release.env')
before = (release / 'release.env.before').read_bytes()
new_main = '{mainapi_img}'
new_front = '{frontend_img}'
compose = ['docker', 'compose', '-p', 'bcai-account', '--env-file', str(release_env), '-f', 'compose.yml', '-f', 'compose.8gb.yml']

def run(args): return subprocess.check_output(args, text=True).strip()
def inspect(name): return json.loads(run(['docker', 'inspect', name]))[0]

def atomic_env(data):
    temp = release_env.with_name('release.env.deploy-tmp')
    with temp.open('xb') as out:
        out.write(data)
        out.flush()
        os.fsync(out.fileno())
    os.chmod(temp, 0o600)
    os.replace(temp, release_env)

def compose_up(services):
    with (release / 'compose-deploy.log').open('ab') as log:
        subprocess.run(compose + ['up', '-d', '--no-deps'] + services, cwd='/opt/bcai-account/deploy', stdout=log, stderr=log, check=True, timeout=180)

def health(service, expected):
    for _ in range(90):
        c = inspect('bcai-account-' + service + '-1')
        assert c['Config']['Image'] == expected
        st = c['State']
        if st.get('Health', {{}}).get('Status') == 'healthy':
            return {{'image': expected, 'health': 'healthy', 'startedat': st['StartedAt']}}
        if not st['Running']:
            raise RuntimeError(service + ' is not running')
        time.sleep(1)
    raise RuntimeError(service + ' did not become healthy')

target = before.decode()
for key, val in [('MAINAPI_IMAGE', new_main), ('FRONTEND_IMAGE', new_front)]:
    target, _ = re.subn(r'(?m)^' + key + r'=[^\\r\\n]*', key + '=' + val, target)

atomic_env(target.encode())

services_to_restart = ['frontend']
if {'True' if deploy_backend else 'False'}:
    services_to_restart = ['mainapi', 'frontend']

compose_up(services_to_restart)

res = {{}}
for s in services_to_restart:
    res[s] = health(s, new_main if s == 'mainapi' else new_front)

(release / 'deploy-result.json').write_text(json.dumps({{
    'deployedat': datetime.datetime.now(datetime.timezone.utc).isoformat(),
    'services': res
}}, indent=2))
print("DEPLOY_OK")
"""
    deploy_res = subprocess.run(
        SSH_BASE + ["python3", "-"],
        input=deploy_script,
        capture_output=True,
        text=True,
        check=True
    )
    if "DEPLOY_OK" not in deploy_res.stdout:
        raise RuntimeError(f"Deploy script error: {deploy_res.stdout}\n{deploy_res.stderr}")
    log(f"✅ Services healthy and running in {time.time() - t0:.1f}s")

    # 7. Live verification
    log("🌐 [Step 6/7] Verifying live endpoints on https://account.bcaicloud.com/...")
    t0 = time.time()
    req = urllib.request.Request("https://account.bcaicloud.com/", headers={"User-Agent": "bcai-deploy-check"})
    with urllib.request.urlopen(req, timeout=15) as r:
        assert r.status == 200, f"Root page returned {r.status}"
    
    # Check auth guard
    try:
        req_gl = urllib.request.Request("https://account.bcaicloud.com/api/gl/accounts", headers={"User-Agent": "bcai-deploy-check"})
        with urllib.request.urlopen(req_gl, timeout=15) as r:
            assert r.status == 401
    except urllib.error.HTTPError as e:
        assert e.code == 401, f"Expected 401 from /api/gl/accounts, got {e.code}"

    log(f"✅ Live endpoints verified in {time.time() - t0:.1f}s")

    # 8. Post-deploy Automated Retention Cleanup (Inquisitive DevOps)
    log("🧹 [Step 7/7] Running remote retention cleanup (pruning images > 72h & build cache)...")
    t0 = time.time()
    try:
        cleanup_cmd = (
            "docker image prune -a --filter 'until=72h' -f >/dev/null 2>&1 && "
            "docker builder prune -f --keep-storage 2GB >/dev/null 2>&1 || true"
        )
        run_ssh(cleanup_cmd)
        log(f"✅ Remote retention cleanup complete in {time.time() - t0:.1f}s (disk space maintained).")
    except Exception as e:
        log(f"⚠️ Warning: Remote retention cleanup encountered an issue: {e}")

    total_elapsed = time.time() - start_total
    log(f"🎉 DEPLOY FINISHED SUCCESSFULLY in {total_elapsed:.1f}s! Release: {tag}")

if __name__ == "__main__":
    main()
