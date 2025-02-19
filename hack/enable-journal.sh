#!/bin/bash

# Copyright OpenFaaS Ltd 2025

# This script is for use with Droplets created on DigitalOcean, it will
# ensure that systemd-journald is configured to log to disk and that
# rsyslog is configured to read from systemd-journald.

# Without this change, no logs will be available in the journal, and only
# /var/log/syslog will be populated.

set -e

echo "Checking systemd-journald logs..."
JOURNAL_STATUS=$(journalctl --no-pager -n 10 2>&1)

if echo "$JOURNAL_STATUS" | grep -q "No journal files were found"; then
    echo "No journal files found. Fixing logging configuration..."
else
    echo "Journald appears to be logging. No changes needed."
    exit 0
fi

# Backup original config before making changes
sudo cp /etc/systemd/journald.conf /etc/systemd/journald.conf.bak

# Ensure Storage is persistent
sudo sed -i '/^#Storage=/c\Storage=persistent' /etc/systemd/journald.conf

# Ensure logs are not forwarded only to syslog
sudo sed -i '/^#ForwardToSyslog=/c\ForwardToSyslog=no' /etc/systemd/journald.conf

# Restart systemd-journald
echo "Restarting systemd-journald..."
sudo systemctl restart systemd-journald

# Check if rsyslog already loads imjournal
if ! grep -q 'module(load="imjournal")' /etc/rsyslog.conf; then
    echo "Adding imjournal module to rsyslog..."
    echo 'module(load="imjournal" StateFile="/var/lib/rsyslog/imjournal.state")' | sudo tee -a /etc/rsyslog.conf
fi

# Restart rsyslog to apply changes
echo "Restarting rsyslog..."
sudo systemctl restart rsyslog

echo "Done. Checking if logs appear in journald..."
journalctl --no-pager -n 10
