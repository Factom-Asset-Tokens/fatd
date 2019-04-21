package cmd

import (
	"github.com/posener/complete"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
)

// Complete runs the CLI completion.
func Complete() bool {
	comp := complete.New("fat-cli", rootCmplCmd)
	return comp.Complete()
}

// generateCmplFlags adds completion for all cmd.Flags() not already present in
// cmplFlags.
func generateCmplFlags(cmd *cobra.Command, cmplFlags complete.Flags) {
	// Due to a bug in cobra.Command.Flags(), we must call LocalFlags()
	// first to get any parent flags merged into cmd.Flags().
	// https://github.com/spf13/cobra/issues/412
	cmd.LocalFlags()
	//fmt.Println("Command:", cmd.Use)
	cmd.Flags().VisitAll(func(flg *flag.Flag) {
		//fmt.Println("Flag:", flg.Name)
		name := "--" + flg.Name
		// If the flag already has a custom completion, there is
		// nothing to do.
		if _, ok := cmplFlags[name]; ok {
			return
		}
		// Add a predictor
		var predict complete.Predictor = complete.PredictAnything
		if flg.Value.Type() == "bool" {
			predict = complete.PredictNothing
		}
		cmplFlags[name] = predict
	})
}

// mergeFlags returns a new complete.Flags that merges all flgs.
func mergeFlags(flgs ...complete.Flags) complete.Flags {
	var size int
	for _, flg := range flgs {
		size += len(flg)
	}
	f := make(complete.Flags, size)
	for _, flg := range flgs {
		for k, v := range flg {
			f[k] = v
		}
	}
	return f
}
