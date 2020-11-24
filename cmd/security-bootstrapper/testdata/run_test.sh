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

# This script is to run both security-boostrapper, listening to a certain port until the installation 
# container signals done, and Vault containers. 
# The Vault container will not start until the go-ahead tcp is signaled.

TEST_NETWORK=test-bootstrapper-net
docker network create --driver bridge $TEST_NETWORK

docker run -d -v /edgex-init:/edgex-init \
  --network $TEST_NETWORK \
  --name edgex-security-installer \
  edgexfoundry/docker-security-bootstrapper-go:0.0.0-dev

docker run -d -p 8200:8200 --entrypoint="/edgex-init/wait_install.sh" \
  -v /edgex-init:/edgex-init --name edgex-vault --cap-add=IPC_LOCK \
  --network $TEST_NETWORK \
  -e 'VAULT_LOCAL_CONFIG={"backend": {"file": {"path": "/vault/file"}}, "default_lease_ttl": "168h", "max_lease_ttl": "720h"}' \
  vault:1.5.3 server

sleep 2
docker logs edgex-security-installer

# give enough time to wait until Vault server is started
sleep 15
docker logs edgex-vault
docker logs edgex-security-installer

sleep 40
# cleanup
docker rm $(docker ps -aq --filter="network=$TEST_NETWORK") -f
docker network rm $TEST_NETWORK
