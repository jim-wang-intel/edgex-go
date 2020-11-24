#!/usr/bin/dumb-init /bin/sh
#  ----------------------------------------------------------------------------------
#  Copyright (c) 2020 Intel Corporation
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

# This is customized entrypoint script for Vault.
# In particular, it waits for the BootstrapPort ready to roll

set -e

echo "Script for waiting scurity install bootstrapping done on the listener"
echo "the current directory is $PWD"

# BOOTSTRAPPER_HOST=
# WAIT_TCP_PORT=

if [ "$1" = 'server' ]; then
  echo "Executing dockerize on vault $@ with waiting on tcp://${BOOTSTRAPPER_HOST}:$WAIT_TCP_PORT"
  /edgex-init/dockerize -wait tcp://${BOOTSTRAPPER_HOST}:$WAIT_TCP_PORT -timeout 60s
  echo 'Starting edgex-vault...'
  exec /usr/local/bin/docker-entrypoint.sh server -log-level=info
fi
