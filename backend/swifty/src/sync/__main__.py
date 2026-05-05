"""CLI entry point. Subcommands wired in later tasks."""
import sys


def main() -> int:
    if len(sys.argv) < 2:
        print("usage: python -m sync {auth|once|install-launchd}", file=sys.stderr)
        return 2
    cmd = sys.argv[1]
    if cmd == "auth":
        from sync.auth import run_auth_flow
        return run_auth_flow()
    if cmd == "once":
        from sync.orchestrator import run_once
        dry_run = "--dry-run" in sys.argv[2:]
        return run_once(dry_run=dry_run)
    if cmd == "install-launchd":
        from sync.installer import install_launchd
        return install_launchd()
    print(f"unknown command: {cmd}", file=sys.stderr)
    return 2


if __name__ == "__main__":
    sys.exit(main())
