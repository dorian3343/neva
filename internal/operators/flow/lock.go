package main

import (
	"github.com/emil14/neva/internal/core"
)

func Lock(io core.IO) error {
	sig, err := io.In.Port("sig")
	if err != nil {
		return err
	}

	dataIn, err := io.In.Port("data")
	if err != nil {
		return err
	}

	dataOut, err := io.Out.Port("data")
	if err != nil {
		return err
	}

	go func() {
		for msg := range dataIn {
			<-sig
			dataOut <- msg
		}
	}()

	return nil
}
