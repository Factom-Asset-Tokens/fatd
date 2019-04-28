package cmd

import (
	goflag "flag"

	"github.com/posener/complete"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
)

var cmpl = complete.New("fat-cli", rootCmplCmd)

var installCompletionFlags = func() *flag.FlagSet {
	cmpl.InstallName = "installcompletion"
	cmpl.UninstallName = "uninstallcompletion"
	// Populate a goflag.FlagSet with the install completion flags.
	goflgs := goflag.NewFlagSet("fat-cli", goflag.ContinueOnError)
	cmpl.AddFlags(goflgs)

	// Create a pflag.FlagSet and copy over the goflag.FlagSet.
	flgs := flag.NewFlagSet("fat-cli", flag.ContinueOnError)
	flgs.AddGoFlagSet(goflgs)
	flgs.MarkHidden("y")

	return flgs
}()

// Complete runs the CLI completion.
func Complete() bool {
	return cmpl.Complete()
}

// generateCmplFlags adds completion for all cmd.Flags() not already present in
// cmplFlags.
func generateCmplFlags(cmd *cobra.Command, cmplFlags complete.Flags) {
	// Due to a bug in cobra.Command.Flags(), we must call LocalFlags()
	// first to get any parent flags merged into cmd.Flags().
	// https://github.com/spf13/cobra/issues/412
	cmd.LocalFlags()
	//errLog.Println("Command:", cmd.Use)
	cmd.Flags().VisitAll(func(flg *flag.Flag) {
		//errLog.Println("Flag:", flg.Name)
		name := "--" + flg.Name
		if flg.Hidden {
			//errLog.Println("hidden")
			delete(cmplFlags, name)
			delete(cmplFlags, "-"+flg.Shorthand)
			return
		}
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
		//errLog.Println("added")
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
