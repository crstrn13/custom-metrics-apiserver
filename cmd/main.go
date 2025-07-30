package main

import (
	"context"
	"net/http"
	"os"
	"time"

	"k8s.io/component-base/logs"
	"k8s.io/component-base/metrics/legacyregistry"
	"k8s.io/klog/v2"
	"github.com/emicklei/go-restful/v3"

	testProvider "crstrn13/mock-backend/pkg/provider"

	"sigs.k8s.io/custom-metrics-apiserver/pkg/apiserver/metrics"
	basecmd "sigs.k8s.io/custom-metrics-apiserver/pkg/cmd"
	"sigs.k8s.io/custom-metrics-apiserver/pkg/provider"
)

type SampleAdapter struct {
	basecmd.AdapterBase

	// Message is printed on successful startup
	Message string
}

func (a *SampleAdapter) makeProviderOrDie() (provider.CustomMetricsProvider, *restful.WebService) {
	client, err := a.DynamicClient()
	if err != nil {
		klog.Fatalf("unable to construct dynamic client: %v", err)
	}

	mapper, err := a.RESTMapper()
	if err != nil {
		klog.Fatalf("unable to construct discovery REST mapper: %v", err)
	}

	return testProvider.NewFakeProvider(client, mapper)
}


func main() {
	logs.InitLogs()
	defer logs.FlushLogs()

	cmd := &SampleAdapter{}
	cmd.Name = "test-adapter"

	cmd.Flags().StringVar(&cmd.Message, "msg", "starting adapter...", "startup message")
	logs.AddFlags(cmd.Flags())
	if err := cmd.Flags().Parse(os.Args); err != nil {
		klog.Fatalf("unable to parse flags: %v", err)
	}

	testProvider, webService := cmd.makeProviderOrDie()
	cmd.WithCustomMetrics(testProvider)

	if err := metrics.RegisterMetrics(legacyregistry.Register); err != nil {
		klog.Fatalf("unable to register metrics: %v", err)
	}

	klog.Infof("%s", cmd.Message)
	// Set up POST endpoint for writing fake metric values
	restful.DefaultContainer.Add(webService)
	go func() {
		// Open port for POSTing fake metrics
		server := &http.Server{
			Addr:              ":8080",
			ReadHeaderTimeout: 3 * time.Second,
		}
		klog.Fatal(server.ListenAndServe())
	}()
	if err := cmd.Run(context.Background()); err != nil {
		klog.Fatalf("unable to run custom metrics adapter: %v", err)
	}
}
