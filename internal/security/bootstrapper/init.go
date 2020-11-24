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
	"context"
	"fmt"
	"sync"

	"github.com/edgexfoundry/edgex-go/internal/security/bootstrapper/container"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/bootstrap/startup"
	"github.com/edgexfoundry/go-mod-bootstrap/di"
)

type Bootstrap struct {
	initNeeded bool
}

func NewBootstrap(initNeeded bool) *Bootstrap {
	return &Bootstrap{
		initNeeded: initNeeded,
	}
}

// BootstrapHandler fulfills the BootstrapHandler contract and performs initialization needed by the data service.
func (b *Bootstrap) BootstrapHandler(_ context.Context, _ *sync.WaitGroup, _ startup.Timer, dic *di.Container) bool {
	conf := container.ConfigurationFrom(dic.Get)
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)

	if b.initNeeded {
		lc.Debug(fmt.Sprintf("init phase done: attempts to start up the listener on bootstrap port: %d",
			conf.StageGate.BootstrapPort))

		go StartListener(conf.StageGate.BootstrapPort, lc)
	}

	// wait on for others to be done: each of tcp dialers is a blocking call
	lc.Debug("Waiting on dependent semaphores required to raise the ready-to-run semaphore ...")
	DialTcp(conf.StageGate.ConsulHost, conf.StageGate.ConsulReadyPort, lc)
	DialTcp(conf.StageGate.PostgresHost, conf.StageGate.PostgresReadyPort, lc)
	DialTcp(conf.StageGate.VaultWorkerHost, conf.StageGate.TokensReadyPort, lc)
	DialTcp(conf.StageGate.RedisHost, conf.StageGate.RedisReadyPort, lc)

	// Reached ready-to-run phase
	lc.Debug(fmt.Sprintf("ready-to-run phase done: attempts to start up the listener on ready-to-run port: %d",
		conf.StageGate.ReadyToRunPort))

	go StartListener(conf.StageGate.ReadyToRunPort, lc)

	return false
}
