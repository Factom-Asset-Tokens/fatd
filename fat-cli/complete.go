package main

import (
	"os"

	"github.com/posener/complete"
)

var (
	Completion = complete.New(os.Args[0], complete.Command{
		Flags: complete.Flags{
			"-debug": complete.PredictNothing,

			"-apiaddress": complete.PredictAnything,

			"-w":              complete.PredictAnything,
			"-wallettimeout":  complete.PredictAnything,
			"-walletuser":     complete.PredictAnything,
			"-walletpassword": complete.PredictAnything,
			"-walletcert":     complete.PredictFiles("*"),
			"-wallettls":      complete.PredictNothing,

			"-s":               complete.PredictAnything,
			"-factomdtimeout":  complete.PredictAnything,
			"-factomduser":     complete.PredictAnything,
			"-factomdpassword": complete.PredictAnything,
			"-factomdcert":     complete.PredictFiles("*"),
			"-factomdtls":      complete.PredictNothing,

			"-y":                   complete.PredictNothing,
			"-installcompletion":   complete.PredictNothing,
			"-uninstallcompletion": complete.PredictNothing,

			"-token":    complete.PredictAnything,
			"-identity": complete.PredictAnything,
			"-chainid":  complete.PredictAnything,
			"-ecpub":    predictAddress(false, 1, "-ecpub", ""),
		},
		Sub: complete.Commands{
			"balance": complete.Command{
				Args: predictAddress(true, 1, "", ""),
			},
			"issue": complete.Command{
				Flags: complete.Flags{
					"-sk1":    complete.PredictAnything,
					"-type":   complete.PredictSet("FAT-0"),
					"-supply": complete.PredictAnything,
					"-symbol": complete.PredictAnything,
					"-name":   complete.PredictAnything,
				},
				Args: complete.PredictAnything,
			},
			"transact": complete.Command{
				Flags: complete.Flags{
					"-sk1":      complete.PredictAnything,
					"-coinbase": complete.PredictAnything,
					"-input": predictAddress(
						true, 1, "-input", ":"),
					"-output": predictAddress(
						true, 1, "-output", ":"),
				},
				Args: complete.PredictAnything,
			},
		},
	})
)
