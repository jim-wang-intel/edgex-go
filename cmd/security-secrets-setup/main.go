/*******************************************************************************
 * Copyright 2019 Dell Inc.
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
 *******************************************************************************/

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/pkg/config"
	"github.com/edgexfoundry/edgex-go/internal/pkg/usage"
	"github.com/edgexfoundry/edgex-go/internal/security/setup"
	"github.com/edgexfoundry/edgex-go/internal/security/setup/certificates"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"

	option "github.com/edgexfoundry/edgex-go/internal/security/pkiinit/option"
)

type exit interface {
	callExit(int)
}

type exitCode struct{}

type optionDispatcher interface {
	run() (int, error)
}

type pkiInitOptionDispatcher struct{}

var exitInstance = newExit()
var dispatcherInstance = newOptionDispatcher()
var helpOpt bool
var generateOpt bool
var importOpt bool
var cacheOpt bool
var cacheCAOpt bool
var configFile string

func init() {
	// define and register command line flags:
	flag.BoolVar(&helpOpt, "h", false, "help message")
	flag.BoolVar(&helpOpt, "help", false, "help message")
	flag.StringVar(&configFile, "config", "", "specify JSON configuration file: /path/to/file.json")
	flag.StringVar(&configFile, "c", "", "specify JSON configuration file: /path/to/file.json")
	flag.BoolVar(&generateOpt, "generate", false, "to generate a new PKI from scratch")
	flag.BoolVar(&importOpt, "import", false, " deploys a PKI from a cached PKI")
	flag.BoolVar(&cacheOpt, "cache", false, "generates a fresh PKI exactly once then caches it for future use")
	flag.BoolVar(&cacheCAOpt, "cacheca", false, "generates a fresh PKI exactly once then caches it with the CA keys presevered")
}

func main() {
	start := time.Now()

	flag.Usage = usage.HelpCallbackSecuritySetup
	flag.Parse()

	if helpOpt {
		// as specified in the requirement, help option terminates with 0 exit status
		flag.Usage()
		exitInstance.callExit(0)
		return
	}

	if configFile == "" {
		flag.PrintDefaults()
		exitInstance.callExit(0)
		return
	}

	if len(os.Args) < 2 {
		fmt.Println("Please specify option for security-secrets-setup.")
		flag.Usage()
		exitInstance.callExit(1)
		return
	}

	statusCode, err := dispatcherInstance.run()

	if err != nil {
		log.Println(err)
	}

	exitInstance.callExit(statusCode)

	setup.Init()

	// Create and initialize the fs environment and global vars for the PKI materials
	lc := logger.NewClient("security-secrets-setup", setup.Configuration.Logging.EnableRemote,
		setup.Configuration.Logging.File, setup.Configuration.Writable.LogLevel)

	// Read the Json config file and unmarshall content into struct type X509Config
	x509config, err := config.NewX509Config(configFile)
	if err != nil {
		lc.Error(err.Error())
		return
	}

	seed, err := setup.NewCertificateSeed(x509config, setup.NewDirectoryHandler(lc))
	if err != nil {
		lc.Error(err.Error())
		return
	}

	rootCA, err := certificates.NewCertificateGenerator(certificates.RootCertificate, seed, certificates.NewFileWriter(), lc)
	if err != nil {
		lc.Error(err.Error())
		return
	}

	err = rootCA.Generate()
	if err != nil {
		lc.Error(err.Error())
		return
	}

	tlsCert, err := certificates.NewCertificateGenerator(certificates.TLSCertificate, seed, certificates.NewFileWriter(), lc)
	if err != nil {
		lc.Error(err.Error())
		return
	}

	tlsCert.Generate()
	if err != nil {
		lc.Error(err.Error())
		return
	}
	lc.Info("PKISetup complete", internal.LogDurationKey, time.Since(start).String())
}

func newExit() exit {
	return &exitCode{}
}

func newOptionDispatcher() optionDispatcher {
	return &pkiInitOptionDispatcher{}
}

func (code *exitCode) callExit(statusCode int) {
	os.Exit(statusCode)
}

func setupPkiInitOption() (executor option.OptionsExecutor, status int, err error) {
	opts := option.PkiInitOption{
		GenerateOpt: generateOpt,
		ImportOpt:   importOpt,
		CacheOpt:    cacheOpt,
		CacheCAOpt:  cacheCAOpt,
	}

	return option.NewPkiInitOption(opts)
}

func (dispatcher *pkiInitOptionDispatcher) run() (statusCode int, err error) {

	optsExecutor, statusCode, err := setupPkiInitOption()
	if err != nil {
		return
	}

	return optsExecutor.ProcessOptions()
}

// TODO: ELIMINATE THIS ----------------------------------------------------------
func fatalIfErr(err error, msg string) {
	if err != nil {
		log.Fatalf("ERROR: %s: %s", msg, err) // fatalf() =  Prinf() followed by a call to os.Exit(1)
	}
}
