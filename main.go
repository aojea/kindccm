package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"
	"net/http"

	"k8s.io/component-base/logs"
	"k8s.io/apiserver/pkg/server/healthz"
	"k8s.io/kubernetes/cmd/cloud-controller-manager/app"
	_ "github.com/aojea/kindccm/kindccm"
)

func init() {
	healthz.InstallHandler(http.DefaultServeMux)
}

func main() {
	rand.Seed(time.Now().UnixNano())

	command := app.NewCloudControllerManagerCommand()
	// TODO: once we switch everything over to Cobra commands, we can go back to calling
	// utilflag.InitFlags() (by removing its pflag.Parse() call). For now, we have to set the
	// normalize func and add the go flag set by hand.

	logs.InitLogs()
	defer logs.FlushLogs()

	if err := command.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
