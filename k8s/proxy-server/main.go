package main

import (
	"k8s.io/klog/v2"
	"sigs.k8s.io/apiserver-runtime/pkg/builder"
)

func main() {
	apiserver := builder.APIServer.
		WithResource(&VM{}).
		WithLocalDebugExtension()

	err := apiserver.Execute()
	if err != nil {
		klog.Fatalf("Failed to start VM API server: %v", err)
	}
}
