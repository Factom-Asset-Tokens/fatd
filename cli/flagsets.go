package main

import flag "github.com/spf13/pflag"

var (
	ecEsAdr      ECEsAddress
	force        bool
	curl         bool
	composeFlags = func() *flag.FlagSet {
		flags := flag.NewFlagSet("", flag.ContinueOnError)
		flags.VarPF(&ecEsAdr, "ecadr", "e",
			"EC or Es address to pay for entries").DefValue = ""
		flags.BoolVar(&force, "force", false,
			"Skip sanity checks for balances, chain status, and sk1 key")
		flags.BoolVar(&curl, "curl", false,
			"Do not submit Factom entry; print curl commands")
		return flags
	}()
)

func whitelistFlags(flags *flag.FlagSet, list ...string) []string {
	var invalid []string
	flags.Visit(func(flg *flag.Flag) {
		var whitelisted bool
		// Compare flg.Name with all whitelisted flags.
		for _, name := range list {
			// Check for very basic globbing.
			if name[len(name)-1] == '*' {
				// Remove the asterisk so that len(name) is
				// correct when used below.
				name = name[:len(name)-1]
			}
			if flg.Name[:min(len(name), len(flg.Name))] == name {
				whitelisted = true
				break
			}
		}
		if whitelisted {
			return
		}
		invalid = append(invalid, flg.Name)
	})
	return invalid
}
func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}
