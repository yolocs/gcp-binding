package main

import (
	"knative.dev/pkg/injection/sharedmain"

	"github.com/yolocs/gcp-binding/pkg/reconciler/backing"
)

func main() {
	sharedmain.Main("controller",
		backing.NewController,
	)
}
