package main

import (
	"fmt"

	flag "github.com/spf13/pflag"
)

func required(set *flag.FlagSet, flags ...string) error {
	var missing string
	for _, f := range flags {
		if !set.Changed(f) {
			missing += " --" + f + ","
		}
	}
	if len(missing) == 0 {
		return nil
	}

	return fmt.Errorf("required:" + missing[:len(missing)-1])
}

func prohibited(set *flag.FlagSet, flags ...string) error {
	var prohibited string
	for _, f := range flags {
		if set.Changed(f) {
			prohibited += " --" + f + ","
		}
	}
	if len(prohibited) == 0 {
		return nil
	}

	return fmt.Errorf("not allowed:" + prohibited[:len(prohibited)-1])
}
