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
set -e

# function to trim off leading and trailing spaces
trim_spaces()
{
  TR=$(echo -e "$1" | sed -e 's/^[[:space:]]*//' -e 's/[[:space:]]*$//')
  echo "${TR}"
}

# Passing the arguments to the executable
if [ "${1:0:1}" = '-' ]; then
    set -- security-bootstrapper "$@"
fi

# get the port numbers from the configuration toml file
TOML_FILE="/edgex/res/configuration.toml"
[[ -e ${TOML_FILE} ]] || { echo "${TOML_FILE} does not exist." >&2;exit 1;}

IFS="="
BOOTSTRAPPER_HOST_KEY="BootstrapperHost"
BOOTSTRAP_PORT_KEY="BootstrapPort"
TOKENS_READY_PORT_KEY="TokensReadyPort"
READY_TO_RUN_PORT_KEY="ReadyToRunPort"
bootstrapper_host=0
bootstrap_port_number=0
tokens_ready_port_number=0
ready_to_run_port_number=0
while read -r key value
do
  trimmed_key=$(trim_spaces "${key}")
  trimmed_value=$(trim_spaces "${value}")

  if [ "${trimmed_key}" = "${BOOTSTRAPPER_HOST_KEY}" ]; then
    bootstrapper_host=${trimmed_value}
  elif [ "${trimmed_key}" = "${BOOTSTRAP_PORT_KEY}" ]; then
    bootstrap_port_number=${trimmed_value}
  elif [ "${trimmed_key}" = "${TOKENS_READY_PORT_KEY}" ]; then
    tokens_ready_port_number=${trimmed_value}
  elif [ "${trimmed_key}" = "${READY_TO_RUN_PORT_KEY}" ]; then
    ready_to_run_port_number=${trimmed_value}
  fi
done < ${TOML_FILE}

echo bootstrapper_host: ${bootstrapper_host}
echo bootstrap_port_number: ${bootstrap_port_number} tokens_ready_port_number: ${tokens_ready_port_number} ready_to_run_port_number:${ready_to_run_port_number}

if [ "$1" = 'security-bootstrapper' ]; then
    echo "Copy dockerize executable"
    cp /usr/local/bin/dockerize /edgex-init/dockerize

    echo 'Injecting edgex-vault entrypoint script...'
    # prepare entrypoint script for Vault
    cp /edgex-staging/vault_wait_install.sh /edgex-staging/tmp.sh
    # replacing the # BOOTSTRAPPER_HOST_KEY= in the entrypoint script with the host name to be used
    sed -i 's/# BOOTSTRAPPER_HOST=/BOOTSTRAPPER_HOST='${bootstrapper_host}'/g' /edgex-staging/tmp.sh
    # replacing the # WAIT_TCP_PORT= in the entrypoint script with the real port number to be used
    sed -i 's/# WAIT_TCP_PORT=/WAIT_TCP_PORT='$bootstrap_port_number'/g' /edgex-staging/tmp.sh
    # update the wait_install script for other services waiting on listener port from installation
    echo "Installing wait_install script"
    cp /edgex-staging/tmp.sh /edgex-init/wait_install.sh
    rm -f /edgex-staging/tmp.sh

    echo 'TBD: injecting consul entrypoint script...'
    # put the preparation of entrypoint script for Consul here
    sleep 1
    echo 'TBD: injecting postgres entrypoint script...'
    # put the preparation of entrypoint script for Postgres here
    sleep 1
    echo 'TBD: injecting redis entrypoint script...'
    # put the preparation of entrypoint script for Redis here
    sleep 1
    echo 'TBD: injecting kong entrypoint script...'
    # put the preparation of entrypoint script for Kong here
    sleep 1

    echo "Executing ./$@"
  "./$@"

else 
  # for debug purposes like docker exec -it security-bootstrapper:0.0.0-dev /bin/sh
  echo "current directory:" $PWD
  exec "$@"
fi

echo "Waiting for termination signal"
exec tail -f /dev/null
