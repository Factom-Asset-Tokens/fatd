package flag

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/Factom-Asset-Tokens/base58"
	"github.com/Factom-Asset-Tokens/fatd/factom"
	"github.com/posener/complete"
	"github.com/sirupsen/logrus"
)

var Revision string

// Environment variable name prefix
const envNamePrefix = "FATD_"

var (
	envNames = map[string]string{
		"startscanheight": "START_SCAN_HEIGHT",
		"debug":           "DEBUG",

		"dbpath": "DB_PATH",

		"apiaddress": "API_ADDRESS",

		"s":              "FACTOMD_SERVER",
		"factomdtimeout": "FACTOMD_TIMEOUT",
		"w":              "WALLETD_SERVER",

		"ecpub": "ECPUB",
	}
	defaults = map[string]interface{}{
		"startscanheight": uint64(0),
		"debug":           false,

		"dbpath": "./fatd.db",

		"apiaddress": ":8078",

		"s":              "localhost:8088",
		"factomdtimeout": time.Duration(0),
		"w":              "localhost:8089",

		"ecpub": "",
	}
	descriptions = map[string]string{
		"startscanheight": "Block height to start scanning for deposits on startup",
		"debug":           "Log debug messages",

		"dbpath": "Path to the folder containing all database files",

		"apiaddress": "IPAddr:port# to bind to for serving the JSON RPC 2.0 API",

		"s":              "IPAddr:port# of factomd API to use to access blockchain",
		"factomdtimeout": "Timeout for factomd API requests, 0 means never timeout",
		"w":              "IPAddr:port# of factom-walletd API to use to access wallet",

		"ecpub": "Entry Credit Public Address to use to pay for Factom entries",
	}
	flags = complete.Flags{
		"-startscanheight": complete.PredictAnything,
		"-debug":           complete.PredictNothing,

		"-dbpath": complete.PredictFiles("*"),

		"-apiaddress": complete.PredictAnything,

		"-s":              complete.PredictAnything,
		"-factomdtimeout": complete.PredictAnything,
		"-w":              complete.PredictAnything,

		"-y":                   complete.PredictNothing,
		"-installcompletion":   complete.PredictNothing,
		"-uninstallcompletion": complete.PredictNothing,

		"-ecpub": predictAddress(false, 1, "-ecpub", ""),
	}

	startScanHeight uint64      // We parse the flag as unsigned.
	StartScanHeight int64  = -1 // We work with the signed value.
	LogDebug        bool

	ECPub string

	DBPath string

	APIAddress string

	FactomClient = &factom.Client{}

	flagset    map[string]bool
	log        *logrus.Entry
	Completion *complete.Complete
)

func init() {
	flagVar(&startScanHeight, "startscanheight")
	flagVar(&LogDebug, "debug")

	flagVar(&DBPath, "dbpath")

	flagVar(&APIAddress, "apiaddress")

	flagVar((*ecpub)(&ECPub), "ecpub")

	flagVar(&FactomClient.FactomdServer, "s")
	flagVar(&FactomClient.Timeout, "factomdtimeout")
	flagVar(&FactomClient.WalletServer, "w")

	// Add flags for self installing the CLI completion tool
	Completion = complete.New(os.Args[0], complete.Command{Flags: flags})
	Completion.CLI.InstallName = "installcompletion"
	Completion.CLI.UninstallName = "uninstallcompletion"
	Completion.AddFlags(nil)
}

func Parse() {
	flag.Parse()
	flagset = make(map[string]bool)
	flag.Visit(func(f *flag.Flag) { flagset[f.Name] = true })

	setupLogger()

	// Load options from environment variables if they haven't been
	// specified on the command line.
	loadFromEnv(&startScanHeight, "startscanheight")
	loadFromEnv(&LogDebug, "debug")

	loadFromEnv(&DBPath, "dbpath")

	loadFromEnv(&APIAddress, "apiaddress")

	loadFromEnv(&FactomClient.FactomdServer, "s")
	loadFromEnv(&FactomClient.Timeout, "factomdtimeout")
	loadFromEnv(&FactomClient.WalletServer, "w")

	loadFromEnv((*ecpub)(&ECPub), "ecpub")

	if flagset["startscanheight"] {
		StartScanHeight = int64(startScanHeight)
	}
}

func Validate() {
	log.Debugf("-dbpath          %#v", DBPath)
	log.Debugf("-apiaddress      %#v", APIAddress)
	log.Debugf("-startscanheight %v ", StartScanHeight)
	debugPrintln()

	log.Debugf("-s              %#v", FactomClient.FactomdServer)
	log.Debugf("-w              %#v", FactomClient.WalletServer)
	log.Debugf("-factomdtimeout %v ", FactomClient.Timeout)
	debugPrintln()
}

func flagVar(v interface{}, name string) {
	dflt := defaults[name]
	desc := description(name)
	switch v := v.(type) {
	case *string:
		flag.StringVar(v, name, dflt.(string), desc)
	case *time.Duration:
		flag.DurationVar(v, name, dflt.(time.Duration), desc)
	case *uint64:
		flag.Uint64Var(v, name, dflt.(uint64), desc)
	case *int64:
		flag.Int64Var(v, name, dflt.(int64), desc)
	case *bool:
		flag.BoolVar(v, name, dflt.(bool), desc)
	case flag.Value:
		flag.Var(v, name, desc)
	}
}

func loadFromEnv(v interface{}, flagName string) {
	if flagset[flagName] {
		return
	}
	eName := envName(flagName)
	eVar, ok := os.LookupEnv(eName)
	if len(eVar) > 0 {
		switch v := v.(type) {
		case flag.Value:
			if err := v.Set(eVar); err != nil {
				log.Fatalf("Environment Variable %v: %v", eName, err)
			}
		case *string:
			*v = eVar
		case *time.Duration:
			duration, err := time.ParseDuration(eVar)
			if err != nil {
				log.Fatalf("Environment Variable %v: "+
					"time.ParseDuration(\"%v\"): %v",
					eName, eVar, err)
			}
			*v = duration
		case *uint64:
			val, err := strconv.ParseUint(eVar, 10, 64)
			if err != nil {
				log.Fatalf("Environment Variable %v: "+
					"strconv.ParseUint(\"%v\", 10, 64): %v",
					eName, eVar, err)
			}
			*v = val
		case *bool:
			if ok {
				*v = true
			}
		}
	}
}

func debugPrintln() {
	if LogDebug {
		fmt.Println()
	}
}

func envName(flagName string) string {
	return envNamePrefix + envNames[flagName]
}
func description(flagName string) string {
	return fmt.Sprintf("%s\nEnvironment variable: %v",
		descriptions[flagName], envName(flagName))
}

func setupLogger() {
	_log := logrus.New()
	_log.Formatter = &logrus.TextFormatter{ForceColors: true,
		DisableTimestamp:       true,
		DisableLevelTruncation: true}
	if LogDebug {
		_log.SetLevel(logrus.DebugLevel)
	}
	log = _log.WithField("pkg", "flag")
}

type ecpub string

// String returns the hex encoded data of b.
func (ec ecpub) String() string {
	return string(ec)
}
func (ec *ecpub) Set(data string) error {
	if len(data) != 52 {
		return fmt.Errorf("invalid length")
	}
	if data[0:2] != "EC" {
		return fmt.Errorf("invalid prefix")
	}
	_, _, err := base58.CheckDecode(data, 2)
	if err != nil {
		return err
	}
	*ec = ecpub(data)
	return nil
}
