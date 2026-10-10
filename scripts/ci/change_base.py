"""Select a published ancestor for push filtering without losing pending changes."""

import json
import os
import subprocess
import urllib.error
import urllib.parse
import urllib.request


def published_base():
    """Return the most recent successful main publication in this commit's history."""
    query = urllib.parse.urlencode({
        "branch": "main", "event": "push", "status": "success", "per_page": 100,
    })
    api = os.environ.get("GITHUB_API_URL", "https://api.github.com")
    repo = os.environ["GITHUB_REPOSITORY"]
    request = urllib.request.Request(
        f"{api}/repos/{repo}/actions/workflows/container.yml/runs?{query}",
        headers={
            "Accept": "application/vnd.github+json",
            "Authorization": f"Bearer {os.environ['GH_TOKEN']}",
            "User-Agent": "tx-carpool-ci",
            "X-GitHub-Api-Version": "2022-11-28",
        },
    )
    with urllib.request.urlopen(request, timeout=30) as response:
        runs = json.load(response)["workflow_runs"]
    for run in runs:
        commit = run["head_sha"]
        ancestor = subprocess.run(
            ["git", "merge-base", "--is-ancestor", commit, "HEAD"],
            check=False, capture_output=True,
        )
        if ancestor.returncode == 0:
            return commit
    return None


def main():
    base = os.environ["GITHUB_REF"]
    full = base.startswith("refs/tags/")
    if os.environ["GITHUB_EVENT_NAME"] == "push" and not full:
        try:
            base = published_base()
        except (OSError, ValueError, KeyError, urllib.error.URLError) as error:
            print(f"Published baseline unavailable ({type(error).__name__}); run all gates.")
            base = None
        full = base is None
    with open(os.environ["GITHUB_OUTPUT"], "a", encoding="utf-8") as output:
        output.write(f"base={base or 'HEAD'}\nfull={str(full).lower()}\n")
    print(f"Change baseline: {base or 'unavailable'}; full validation: {full}")


if __name__ == "__main__":
    main()
