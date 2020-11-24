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
	"fmt"
	"net"
	"os"
	"time"

	"github.com/edgexfoundry/edgex-go/internal/security/secrets/contract"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
)

// StartListener will instantiate a new listener that signals process is done
func StartListener(port int, lc logger.LoggingClient) {
	lc.Info(fmt.Sprintf("Starting listener on port %d to signal the process is done...\n", port))

	doneSrv := fmt.Sprintf(":%d", port)

	listener, err := net.Listen("tcp", doneSrv)
	checkError(err)

	for {
		conn, err := listener.Accept()
		if err != nil {
			lc.Error(fmt.Sprintf("found error %v when accepting connection!\n", err))
			time.Sleep(2 * time.Second)
			continue
		}

		go func() {
			// raise the semaphore on port
			daytime := time.Now().String()
			_, _ = conn.Write([]byte(daytime)) // don't care about return value
			_ = conn.Close()
			// intended process listener is done
			lc.Info(fmt.Sprintf("listener on port %d is done", port))
		}()
	}
}

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s\n", err.Error())
		os.Exit(contract.StatusCodeExitWithError)
	}
}
