#!/usr/bin/dumb-init /bin/sh
#  ----------------------------------------------------------------------------------
#  Copyright (c) 2019 Intel Corporation
#
#  Licensed under the Apache License, Version 2.0 (the "License");
#  you may not use this file except in compliance with the License.
#  You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
#   Unless required by applicable law or agreed to in writing, software
#   distributed under the License is distributed on an "AS IS" BASIS,
#   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#   See the License for the specific language governing permissions and
#   limitations under the License.
#
#  SPDX-License-Identifier: Apache-2.0'
#  ----------------------------------------------------------------------------------

set -e

# Use dumb-init as PID 1 in order to reap zombie processes and forward system signals to 
# all processes in its session. This can alleviate the chance of leaking zombies, 
# thus more graceful termination of all sub-processes if any.

# runtime directory is set per user:
XDG_RUNTIME_DIR=${XDG_RUNTIME_DIR:-/run/user/$(echo $(id -u))}
PATH="$BASE_DIR:$PATH"
export XDG_RUNTIME_DIR PATH

# debug output:
echo XDG_RUNTIME_DIR $XDG_RUNTIME_DIR
echo BASE_DIR $BASE_DIR

# if running security-secrets-setup subcommand
# build full command line into positional args
if [ "$1" = 'generate' -o "$1" = 'cache' -o "$1" = 'import' -o "$1" = 'legacy' ]; then
    set -- security-secrets-setup "$@"
fi

posthook=""
if [ "$1" = 'security-secrets-setup' ]; then
    posthook="chown -R 100:1000 /run/edgex/secrets/edgex-vault/ /vault/init/"
fi

echo "Executing $@"
"$@"
echo "Executing hook=$posthook"
$posthook
