package main

import (
	"fmt"
	"os"

	"github.com/golang/glog"
	"github.com/spf13/pflag"

	"k8s.io/apiserver/pkg/server/healthz"
	cloudprovider "k8s.io/cloud-provider"
	"k8s.io/component-base/cli/flag"
	"k8s.io/component-base/logs"
	"k8s.io/kubernetes/cmd/cloud-controller-manager/app"
	"k8s.io/kubernetes/cmd/cloud-controller-manager/app/options"

	"github.com/aojea/kindccm/kindccm"
)

const version = "1.0.0"

func init() {
	healthz.DefaultHealthz()
}

func main() {
	s := options.NewCloudControllerManagerServer()
	s.AddFlags(pflag.CommandLine)
	addVersionFlag()

	flag.InitFlags()
	logs.InitLogs()
	defer logs.FlushLogs()

	printAndExitIfRequested()

	cloud, err := cloudprovider.InitCloudProvider(kindccm.ProviderName, s.CloudConfigFile)
	if err != nil {
		glog.Fatalf("Cloud provider could not be initialized: %v", err)
	}

	glog.Info("Starting version ", version)
	if err := app.Run(s, cloud); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
