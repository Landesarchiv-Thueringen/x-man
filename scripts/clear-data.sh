#!/usr/bin/env bash

set -e
cd -- "$( dirname -- "${BASH_SOURCE[0]}" )/.."
git clean -dxf data/archive data/message_store data/mongo data/transfer_dir data/webdav