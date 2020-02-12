#!/bin/sh
# docker-2ndboot.sh is executed by docker-entrypoint.sh
set -eux
systemctl start containerd
faasd install
exec journalctl -f
