#!/usr/bin/env python3
"""E2E test runner for Claude Code WebUI.

Usage:
    uv run main.py                   # run all test cases
    uv run main.py --case basic_chat # run a single test case
"""

import argparse
import importlib
import logging
import os
import signal
import subprocess
import sys
import time
from datetime import datetime
from pathlib import Path

from playwright.sync_api import sync_playwright

PROJECT_ROOT = Path(__file__).resolve().parent.parent.parent
DATA_DIR = PROJECT_ROOT / "data" / "e2e" / "cases"
BACKEND_BIN = PROJECT_ROOT / "data" / "rpc" / "rpc"
BACKEND_PORT = "18080"
BASE_URL = f"http://localhost:{BACKEND_PORT}"


def discover_cases(cases_dir: Path) -> list[str]:
    """Find all test_*.py modules in the cases/ directory."""
    results = []
    for f in sorted(cases_dir.glob("test_*.py")):
        results.append(f.stem)
    return results


def setup_output_dir(case_name: str) -> Path:
    """Create timestamped output directory for a test case."""
    ts = datetime.now().strftime("%Y%m%d-%H%M%S")
    out = DATA_DIR / f"{ts}-{case_name}"
    (out / "screenshots").mkdir(parents=True, exist_ok=True)
    (out / "logs").mkdir(parents=True, exist_ok=True)
    return out


def setup_logger(log_file: Path) -> logging.Logger:
    logger = logging.getLogger("e2e")
    logger.setLevel(logging.DEBUG)

    for h in logger.handlers[:]:
        logger.removeHandler(h)

    fh = logging.FileHandler(log_file, mode="w")
    fh.setLevel(logging.DEBUG)
    fh.setFormatter(logging.Formatter("%(asctime)s [%(levelname)s] %(message)s"))
    logger.addHandler(fh)

    sh = logging.StreamHandler()
    sh.setLevel(logging.INFO)
    sh.setFormatter(logging.Formatter("%(asctime)s [%(levelname)s] %(message)s"))
    logger.addHandler(sh)

    return logger


def start_backend(logger: logging.Logger, out_dir: Path | None = None) -> subprocess.Popen:
    """Start the backend server process."""
    env = os.environ.copy()
    env["PORT"] = BACKEND_PORT

    logger.info(f"Starting backend on port {BACKEND_PORT}...")
    log_path = (out_dir / "logs" / "backend.log") if out_dir else None
    if log_path:
        log_file = open(log_path, "w")
    else:
        log_file = open(os.devnull, "w")

    proc = subprocess.Popen(
        [str(BACKEND_BIN)],
        cwd=str(PROJECT_ROOT),
        env=env,
        stdin=subprocess.DEVNULL,
        stdout=log_file,
        stderr=subprocess.STDOUT,
    )
    proc._backend_log_file = log_file  # type: ignore[attr-defined]
    return proc


def wait_for_backend(logger: logging.Logger, timeout: float = 10.0) -> bool:
    """Wait until the backend is accepting connections."""
    import urllib.request
    import urllib.error

    deadline = time.time() + timeout
    while time.time() < deadline:
        try:
            url = f"{BASE_URL}/pkg.rpc.ClaudeService/Healthz"
            req = urllib.request.Request(
                url,
                data=b"{}",
                headers={"Content-Type": "application/json"},
                method="POST",
            )
            urllib.request.urlopen(req, timeout=2)
            logger.info("Backend is ready")
            return True
        except (urllib.error.URLError, ConnectionRefusedError, OSError):
            time.sleep(0.3)
    logger.error("Backend failed to start within timeout")
    return False


def stop_backend(proc: subprocess.Popen, logger: logging.Logger):
    """Stop the backend process gracefully."""
    if proc.poll() is None:
        logger.info("Stopping backend...")
        proc.send_signal(signal.SIGTERM)
        try:
            proc.wait(timeout=5)
        except subprocess.TimeoutExpired:
            proc.kill()
            proc.wait()
        logger.info("Backend stopped")
    if hasattr(proc, "_backend_log_file"):
        proc._backend_log_file.close()  # type: ignore[attr-defined]


def run_case(case_name: str, module_path: Path) -> bool:
    """Run a single test case. Returns True if passed."""
    out_dir = setup_output_dir(case_name)
    logger = setup_logger(out_dir / "logs" / "test.log")

    logger.info(f"=== Running test case: {case_name} ===")

    backend_proc = start_backend(logger, out_dir)
    try:
        if not wait_for_backend(logger):
            logger.error("Backend not ready, skipping test")
            return False

        with sync_playwright() as pw:
            browser = pw.chromium.launch(headless=True)
            context = browser.new_context(
                viewport={"width": 1280, "height": 720},
            )
            context.tracing.start(screenshots=True, snapshots=True)

            page = context.new_page()

            spec = importlib.import_module(f"cases.{case_name}")
            try:
                passed = spec.run(page, BASE_URL, out_dir, logger)
            except Exception as e:
                logger.exception(f"Test case failed with exception: {e}")
                page.screenshot(path=str(out_dir / "screenshots" / "error.png"))
                passed = False

            trace_path = out_dir / "logs" / "trace.zip"
            context.tracing.stop(path=str(trace_path))
            logger.info(f"Trace saved to {trace_path}")

            browser.close()

    finally:
        stop_backend(backend_proc, logger)

    status = "PASSED" if passed else "FAILED"
    logger.info(f"=== {case_name}: {status} ===")
    return passed


def main():
    parser = argparse.ArgumentParser(description="E2E test runner")
    parser.add_argument("--case", help="Run a specific test case")
    args = parser.parse_args()

    cases_dir = Path(__file__).parent / "cases"

    if not BACKEND_BIN.exists():
        print(f"Error: Backend binary not found at {BACKEND_BIN}", file=sys.stderr)
        print("Run 'task build:bin' first.", file=sys.stderr)
        sys.exit(1)

    if args.case:
        case_names = [args.case if args.case.startswith("test_") else f"test_{args.case}"]
    else:
        case_names = discover_cases(cases_dir)

    if not case_names:
        print("No test cases found in", cases_dir, file=sys.stderr)
        sys.exit(1)

    print(f"Running {len(case_names)} test case(s)...")
    results: dict[str, bool] = {}

    for name in case_names:
        results[name] = run_case(name, cases_dir / f"{name}.py")

    print("\n" + "=" * 50)
    print("Results:")
    all_passed = True
    for name, passed in results.items():
        status = "PASS" if passed else "FAIL"
        print(f"  {status}  {name}")
        if not passed:
            all_passed = False

    print("=" * 50)
    if all_passed:
        print(f"All {len(results)} test(s) passed!")
    else:
        failed = sum(1 for p in results.values() if not p)
        print(f"{failed}/{len(results)} test(s) failed!")
        sys.exit(1)


if __name__ == "__main__":
    main()
