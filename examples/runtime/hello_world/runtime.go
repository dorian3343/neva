package main

import (
	"log"

	"github.com/emil14/neva/internal/runtime"
	"github.com/emil14/neva/internal/runtime/builder"
	"github.com/emil14/neva/internal/runtime/decoder"
	"github.com/emil14/neva/internal/runtime/executor"
	"github.com/emil14/neva/internal/runtime/executor/connector"
	logginginterceptor "github.com/emil14/neva/internal/runtime/executor/connector/interceptor/log"
	"github.com/emil14/neva/internal/runtime/executor/effector"
	operatorseffects "github.com/emil14/neva/internal/runtime/executor/effector/component"
	constantseffects "github.com/emil14/neva/internal/runtime/executor/effector/constant"
	oprepo "github.com/emil14/neva/internal/runtime/executor/effector/operator/repo"
	triggerseffects "github.com/emil14/neva/internal/runtime/executor/effector/trigger"
)

func mustCreateRuntime() runtime.Runtime {
	l := log.Default()
	l.SetFlags(log.Flags() &^ (log.Ldate | log.Ltime))

	r := runtime.MustNew(
		decoder.MustNewProto(
			decoder.NewCaster(),
			decoder.NewUnmarshaler(),
		),
		builder.Builder{},
		executor.MustNew(
			effector.MustNew(
				constantseffects.Effector{},
				operatorseffects.MustNewEffector(
					oprepo.NewPlugin(map[string]oprepo.File{
						"io": {
							Path:    "/home/evaleev/projects/neva/plugins/io.so",
							Exports: []string{"Println"},
						},
					}),
				),
				triggerseffects.Effector{},
			),
			connector.MustNew(
				logginginterceptor.MustNew(l),
			),
		),
	)

	return r
}
