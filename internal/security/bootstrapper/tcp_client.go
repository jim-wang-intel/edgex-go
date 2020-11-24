/*******************************************************************************
 * Copyright 2020 Intel Corporation
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License
 * is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
 * or implied. See the License for the specific language governing permissions and limitations under
 * the License.
 *
 *******************************************************************************/

package bootstrapper

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
)

// DialTcp will instantiate a new TCP dialer trying to connect to the TCP server specified by host and port
func DialTcp(host string, port int, lc logger.LoggingClient) {
	tcpHost := strings.TrimSpace(host)
	if len(tcpHost) == 0 || port == 0 {
		// log and ignore those
		lc.Info(fmt.Sprintf("Skipping: no TCP server specified, host=%s, port=%d", host, port))
		return
	}

	lc.Info(fmt.Sprintf("Trying to connecting to TCP server %s on port %d", tcpHost, port))

	tcpServerAddr := net.JoinHostPort(host, strconv.Itoa(port))

	for { // keep trying until server connects
		c, err := net.Dial("tcp", tcpServerAddr)
		if err != nil {
			lc.Debug(fmt.Sprintf("TCP server %s is not available yet, retry in 1 second", tcpServerAddr))
			time.Sleep(time.Second)
			continue
		}
		defer func() {
			_ = c.Close()
		}()

		// connected: read the 1st line if anything
		message, _ := bufio.NewReader(c).ReadString('\n')
		lc.Debug(fmt.Sprintf("message from server: %s", message))
		lc.Info(fmt.Sprintf("Connected with TCP server %s ", tcpServerAddr))
		break
	}
}
