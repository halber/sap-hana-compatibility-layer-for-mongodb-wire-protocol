// Copyright 2021 FerretDB Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/expfmt"
	"go.uber.org/zap"
	"golang.org/x/sys/unix"

	"github.com/SAP/sap-hana-compatibility-layer-for-mongodb-wire-protocol/internal/clientconn"
	"github.com/SAP/sap-hana-compatibility-layer-for-mongodb-wire-protocol/internal/hana"
	"github.com/SAP/sap-hana-compatibility-layer-for-mongodb-wire-protocol/internal/handlers"
	"github.com/SAP/sap-hana-compatibility-layer-for-mongodb-wire-protocol/internal/util/debug"
	"github.com/SAP/sap-hana-compatibility-layer-for-mongodb-wire-protocol/internal/util/logging"
	"github.com/SAP/sap-hana-compatibility-layer-for-mongodb-wire-protocol/internal/util/version"
)

//nolint:gochecknoglobals // flags are defined there to be visible in `bin/SAPHANACompatibilitylayer-testcover -h` output
var (
	debugAddrF       = flag.String("debug-addr", "127.0.0.1:8088", "debug address")
	listenAddrF      = flag.String("listen-addr", "127.0.0.1:27017", "listen address")
	modeF            = flag.String("mode", string(clientconn.AllModes[0]), fmt.Sprintf("operation mode: %v", clientconn.AllModes))
	proxyAddrF       = flag.String("proxy-addr", "127.0.0.1:37017", "")
	tlsF             = flag.Bool("tls", false, "enable insecure TLS")
	versionF         = flag.Bool("version", false, "print version to stdout (full version, commit, branch, dirty flag) and exit")
	testConnTimeoutF = flag.Duration("test-conn-timeout", 0, "test: set connection timeout")
	saphanaURL       = flag.String("saphana-url", "please-insert-sap-hana-url-here", "SAP HANA URL")
)

func main() {
	logging.Setup(zap.DebugLevel)
	logger := zap.L()
	flag.Parse()

	info := version.Get()

	if *versionF {
		fmt.Fprintln(os.Stdout, "version:", info.Version)
		fmt.Fprintln(os.Stdout, "commit:", info.Commit)
		fmt.Fprintln(os.Stdout, "branch:", info.Branch)
		fmt.Fprintln(os.Stdout, "dirty:", info.Dirty)
		return
	}

	logger.Info(
		"Starting SAP HANA compatibility layer for MongoDB Wire Protocol "+info.Version+"...",
		zap.String("version", info.Version),
		zap.String("commit", info.Commit),
		zap.String("branch", info.Branch),
		zap.Bool("dirty", info.Dirty),
	)

	var found bool
	for _, m := range clientconn.AllModes {
		if *modeF == string(m) {
			found = true
			break
		}
	}
	if !found {
		logger.Sugar().Fatalf("Unknown mode %q.", *modeF)
	}

	if *tlsF {
		logger.Sugar().Warn("The current TLS implementation is not secure.")
	}

	ctx, stop := signal.NotifyContext(context.Background(), unix.SIGTERM, unix.SIGINT)
	go func() {
		<-ctx.Done()
		logger.Info("Stopping...")
		stop()
	}()

	go debug.RunHandler(ctx, *debugAddrF, logger.Named("debug"))

	hanaPool, err := hana.CreatePool(*saphanaURL, logger, false)
	if err != nil {
		logger.Fatal(err.Error())
	}

	defer hanaPool.Close()

	listenerMetrics := clientconn.NewListenerMetrics()
	handlersMetrics := handlers.NewMetrics()
	prometheus.DefaultRegisterer.MustRegister(listenerMetrics, handlersMetrics)

	l := clientconn.NewListener(&clientconn.NewListenerOpts{
		ListenAddr:      *listenAddrF,
		TLS:             *tlsF,
		ProxyAddr:       *proxyAddrF,
		Mode:            clientconn.Mode(*modeF),
		HanaPool:        hanaPool,
		Logger:          logger.Named("listener"),
		Metrics:         listenerMetrics,
		HandlersMetrics: handlersMetrics,
		TestConnTimeout: *testConnTimeoutF,
	})

	err = l.Run(ctx)
	if err == nil || err == context.Canceled {
		logger.Info("Listener stopped")
	} else {
		logger.Error("Listener stopped", zap.Error(err))
	}

	mfs, err := prometheus.DefaultGatherer.Gather()
	if err != nil {
		panic(err)
	}
	for _, mf := range mfs {
		if _, err := expfmt.MetricFamilyToText(os.Stderr, mf); err != nil {
			panic(err)
		}
	}
}
