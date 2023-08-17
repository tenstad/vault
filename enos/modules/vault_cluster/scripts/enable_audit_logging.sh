#!/bin/env sh
# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: BUSL-1.1


set -eux

LOG_DIR="$(dirname "$LOG_FILE_PATH")"

# Run ncat to listen to port 9090
ncat -l 9090 > "$LOG_DIR/socket.log" &
NCAT_PID=$!

$VAULT_BIN_PATH audit enable file file_path="$LOG_FILE_PATH"
$VAULT_BIN_PATH audit enable syslog tag="vault" facility="AUTH"
$VAULT_BIN_PATH audit enable socket address="127.0.0.1:9090"
