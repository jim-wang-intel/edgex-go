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
	"os"

	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/security/bootstrapper/config"
	"github.com/edgexfoundry/edgex-go/internal/security/bootstrapper/container"

	"github.com/edgexfoundry/go-mod-bootstrap/bootstrap"
	"github.com/edgexfoundry/go-mod-bootstrap/bootstrap/flags"
	"github.com/edgexfoundry/go-mod-bootstrap/bootstrap/interfaces"
	"github.com/edgexfoundry/go-mod-bootstrap/bootstrap/startup"
	"github.com/edgexfoundry/go-mod-bootstrap/di"

	"github.com/gorilla/mux"
)

const (
	securityBootstrapperServiceKey = "edgex-security-bootstrapper"
)

func Main(ctx context.Context, cancel context.CancelFunc, _ *mux.Router, _ chan<- bool) {
	// service key for this bootstrapper service
	startupTimer := startup.NewStartUpTimer(securityBootstrapperServiceKey)

	var initNeeded bool

	f := flags.NewWithUsage(
		"    --init=true/false Indicates if bootstrapper should be initialized again",
	)

	if len(os.Args) < 2 {
		f.Help()
	}

	f.FlagSet.BoolVar(&initNeeded, "init", false, "")
	f.Parse(os.Args[1:])

	configuration := &config.ConfigurationStruct{}
	dic := di.NewContainer(di.ServiceConstructorMap{
		container.ConfigurationName: func(get di.Get) interface{} {
			return configuration
		},
	})

	serviceHandler := NewBootstrap(initNeeded)

	bootstrap.Run(
		ctx,
		cancel,
		f,
		securityBootstrapperServiceKey,
		internal.ConfigStemSecurity+internal.ConfigMajorVersion,
		configuration,
		startupTimer,
		dic,
		[]interfaces.BootstrapHandler{
			serviceHandler.BootstrapHandler,
		},
	)
}
