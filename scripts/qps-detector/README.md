# QPS detector node reporter

- `install.sh` installs or removes the Debian 13 locked cron reporter, keeps its node token root-only, and verifies an initial authenticated report before returning success.
- The reporter stores an inode and byte cursor for Remnawave's Xray output, reads only new complete lines on each interval, handles truncation or rotation, and uploads privacy-sensitive logs in 15 MiB bounded batches without a continuous log-tail process.
