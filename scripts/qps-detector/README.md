# QPS detector node reporter

- `install.sh` installs or removes the Debian 13 locked cron reporter, keeps its node token root-only, and verifies an initial authenticated report before returning success.
- `bash install.sh test` prints up to 1,000 raw Xray core log lines through `remnanode xlogs`, falling back to `docker exec remnanode xlogs`, without installing, uploading, or changing reporter state.
- The reporter stores an inode and byte cursor for Remnawave's Xray output, reads only new complete lines on each interval, handles a not-yet-created log plus later truncation or rotation, and uploads privacy-sensitive logs in 15 MiB bounded batches without a continuous log-tail process.
