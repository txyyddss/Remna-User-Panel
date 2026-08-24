# QPS detector node reporter

- `install.sh` installs or removes the Debian 13 locked cron reporter, keeps its node token root-only, and verifies an initial authenticated report before returning success.
- `bash install.sh test` prints up to 1,000 raw Xray core log lines through `remnanode xlogs`, falling back to `docker exec remnanode xlogs`, without installing, uploading, or changing reporter state.
- The reporter takes a five-second bounded `xlogs` snapshot, keeps the newest 1,000 complete Xray lines, and uploads them in privacy-sensitive 15 MiB batches; server-side fingerprints discard repeated lines between intervals.
