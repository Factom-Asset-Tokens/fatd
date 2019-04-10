package flag

import (
	"flag"
	"os"
	"strings"
	"time"

	"github.com/AdamSLevy/factom"
	"github.com/posener/complete"
)

func predictAddress(fa bool, num int, flagName, suffix string) complete.PredictFunc {
	if len(flagName) == 0 {
		return func(a complete.Args) []string {
			// Count the number of complete arguments that are not flags.
			argc := len(a.Completed[1:])
			for _, arg := range a.Completed[1:] {
				if string(arg[0]) == "-" {
					argc--
				}
			}
			if len(suffix) > 0 && len(a.Last) > 0 &&
				a.Last[len(a.Last)-1:len(a.Last)] == suffix {
				return nil
			}
			if argc < num {
				adrs := listAddresses(fa)
				if len(suffix) > 0 {
					for i := range adrs {
						adrs[i] += suffix
					}
				}
				return adrs
			}
			return nil
		}
	}
	return func(a complete.Args) []string {
		// Count the number of complete arguments that are not flags.
		argc := 0
		for i := len(a.Completed) - 1; i > 0; i-- {
			arg := a.Completed[i]
			if string(arg) == flagName {
				break
			}
			argc++
		}
		if len(suffix) > 0 && len(a.Last) > 0 &&
			a.Last[len(a.Last)-1:len(a.Last)] == suffix {
			return nil
		}
		if argc < num {
			adrs := listAddresses(fa)
			if len(suffix) > 0 {
				for i := range adrs {
					adrs[i] += suffix
				}
			}
			return adrs
		}
		return nil
	}
}

func listAddresses(fa bool) []string {
	fcts, ecs := addressPubStrings()
	if fa {
		return fcts
	}
	return ecs
}

func addressPubStrings() ([]string, []string) {
	parseWalletFlags()
	// Fetch all addresses.
	fcts, ecs, err := factom.FetchAddresses()
	if err != nil {
		complete.Log("error: %v", err)
		return nil, nil
	}

	// Create slices of the public address strings.
	fctAddresses := make([]string, len(fcts))
	for i, fct := range fcts {
		fctAddresses[i] = fct.String()
	}
	ecAddresses := make([]string, len(ecs))
	for i, ec := range ecs {
		ecAddresses[i] = ec.PubString()
	}
	return fctAddresses, ecAddresses
}

var cliFlags *flag.FlagSet

// Parse any previously specified factom-cli options required for connecting to
// factom-walletd
func parseWalletFlags() {
	if cliFlags != nil {
		// We already parsed the flags.
		return
	}
	// Using flag.FlagSet allows us to parse a custom array of flags
	// instead of this programs args.
	cliFlags = flag.NewFlagSet("", flag.ContinueOnError)
	cliFlags.StringVar(&factom.RpcConfig.WalletServer, "w", "localhost:8089", "")
	cliFlags.StringVar(&factom.RpcConfig.WalletTLSCertFile, "walletcert",
		"~/.factom/walletAPIpub.cert", "")
	cliFlags.StringVar(&factom.RpcConfig.WalletRPCUser, "walletuser", "", "")
	cliFlags.StringVar(&factom.RpcConfig.WalletRPCPassword, "walletpassword", "", "")
	cliFlags.BoolVar(&factom.RpcConfig.WalletTLSEnable, "wallettls", false, "")

	// flags.Parse will print warnings if it comes across an unrecognized
	// flag. We don't want this so we temprorarily redirect everything to
	// /dev/null before we call flags.Parse().
	stdout := os.Stdout
	stderr := os.Stderr
	os.Stdout, _ = os.Open(os.DevNull)
	os.Stderr = os.Stdout

	// The current command line being typed is stored in the environment
	// variable COMP_LINE. We split on spaces and discard the first in the
	// list because it is the program name `factom-cli`.
	cliFlags.Parse(strings.Fields(os.Getenv("COMP_LINE"))[1:])

	// Restore stdout and stderr.
	os.Stdout = stdout
	os.Stderr = stderr

	// We want need factom-walletd to timeout or the CLI completion will
	// hang and never return. This is the whole reason we use AdamSLevy's
	// fork of factom.
	factom.SetWalletTimeout(1 * time.Second)
}
