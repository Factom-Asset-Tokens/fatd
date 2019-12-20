// MIT License
//
// Copyright 2018 Canonical Ledgers, LLC
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to
// deal in the Software without restriction, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense, and/or
// sell copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS
// IN THE SOFTWARE.

package flag

import (
	"context"
	"flag"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/Factom-Asset-Tokens/factom"
	"github.com/posener/complete"
	"github.com/sirupsen/logrus"
)

var Revision string

// Environment variable name prefix
const envNamePrefix = "FATD_"

var (
	envNames = map[string]string{
		"startscanheight":    "START_SCAN_HEIGHT",
		"factomscanretries":  "FACTOM_SCAN_RETRIES",
		"factomscaninterval": "FACTOM_SCAN_INTERVAL",
		"debug":              "DEBUG",
		"disablepending":     "DISABLE_PENDING",

		"dbpath": "DB_PATH",

		"apiaddress":  "API_ADDRESS",
		"apiusername": "API_USERNAME",
		"apipassword": "API_PASSWORD",
		"apitlscert":  "API_TLS_CERT",
		"apitlskey":   "API_TLS_KEY",
		"apimaxlimit": "API_MAX_LIMIT",
		"apitimeout":  "API_TIMEOUT",

		"s":               "FACTOMD_SERVER",
		"factomdtimeout":  "FACTOMD_TIMEOUT",
		"factomduser":     "FACTOMD_USER",
		"factomdpassword": "FACTOMD_PASSWORD",
		//"factomdcert":     "FACTOMD_TLS_CERT",
		//"factomdtls":      "FACTOMD_TLS_ENABLE",

		"w":              "WALLETD_SERVER",
		"wallettimeout":  "WALLETD_TIMEOUT",
		"walletuser":     "WALLETD_USER",
		"walletpassword": "WALLETD_PASSWORD",
		//"walletcert":     "WALLETD_TLS_CERT",
		//"wallettls":      "WALLETD_TLS_ENABLE",

		"ecadr": "ECADR",
		"esadr": "ESADR",

		"networkid": "NETWORK_ID",

		"whitelist":        "WHITELIST",
		"blacklist":        "BLACKLIST",
		"ignorenewchains":  "IGNORE_NEW_CHAINS",
		"skipdbvalidation": "SKIP_DB_VALIDATION",
	}
	defaults = map[string]interface{}{
		"startscanheight":    uint64(0),
		"factomscanretries":  int64(0),
		"factomscaninterval": 15 * time.Second,
		"debug":              false,
		"disablepending":     false,

		"dbpath": func() string {
			if home, err := os.UserHomeDir(); err == nil {
				return home + "/.fatd"
			}
			return "./fatd.db"
		}(),

		"apiaddress":  ":8078",
		"apiusername": "",
		"apipassword": "",
		"apitlscert":  "",
		"apitlskey":   "",
		"apimaxlimit": uint64(math.MaxUint32),
		"apitimeout":  5 * time.Second,

		"s":               "http://localhost:8088/v2",
		"factomdtimeout":  20 * time.Second,
		"factomduser":     "",
		"factomdpassword": "",
		//"factomdcert":     "",
		//"factomdtls":      false,

		"w":              "http://localhost:8089/v2",
		"wallettimeout":  10 * time.Second,
		"walletuser":     "",
		"walletpassword": "",
		//"walletcert":     "",
		//"wallettls":      false,

		"ecadr": "",
		"esadr": "",

		"ignorenewchains":  false,
		"skipdbvalidation": false,
	}
	descriptions = map[string]string{
		"startscanheight":    "Block height to start scanning for deposits on startup",
		"factomscanretries":  "Number of times to consecutively retry fetching the latest height before exiting, use -1 for unlimited",
		"factomscaninterval": "Scan interval for new blocks or pending entries",
		"debug":              "Log debug messages",
		"disablepending":     "Do not scan for pending txs, reducing memory usage",

		"dbpath": "Path to the folder containing all database files",

		"apiaddress":  "IPAddr:port# to bind to for serving the fatd API",
		"apiusername": "Username required for connections to fatd API",
		"apipassword": "Password required for connections to fatd API",
		"apitlscert":  "Path to TLS certificate for the fatd API",
		"apitlskey":   "Path to TLS Key for the fatd API",
		"apimaxlimit": "Maximum pagination limit",
		"apitimeout":  "Maximum amount of time to allow API queries to complete",

		"s":               "IPAddr:port# of factomd API to use to access blockchain",
		"factomdtimeout":  "Timeout for factomd API requests, 0 means never timeout",
		"factomduser":     "Username for API connections to factomd",
		"factomdpassword": "Password for API connections to factomd",
		//"factomdcert":     "The TLS certificate that will be provided by the factomd API server",
		//"factomdtls":      "Set to true to use TLS when accessing the factomd API",
		"networkid": `Accepts "main", "test", "localnet", or four bytes in hex`,

		"w":              "IPAddr:port# of factom-walletd API to use to access wallet",
		"wallettimeout":  "Timeout for factom-walletd API requests, 0 means never timeout",
		"walletuser":     "Username for API connections to factom-walletd",
		"walletpassword": "Password for API connections to factom-walletd",
		//"walletcert":     "The TLS certificate that will be provided by the factom-walletd API server",
		//"wallettls":      "Set to true to use TLS when accessing the factom-walletd API",

		"ecadr": "Entry Credit Public Address to use to pay for Factom entries",
		"esadr": "Entry Credit Secret Address to use to pay for Factom entries",

		"whitelist":        "Track only these chains, creating the database if needed",
		"blacklist":        "Do not track or sync these chains, overrides -whitelist",
		"ignorenewchains":  "Do not track new chains, sync existing chain databases",
		"skipdbvalidation": "Skip the full validation check of all chain databases",
	}
	flags = complete.Flags{
		"-startscanheight":    complete.PredictAnything,
		"-factomscanretries":  complete.PredictAnything,
		"-factomscaninterval": complete.PredictAnything,
		"-debug":              complete.PredictNothing,
		"-disablepending":     complete.PredictNothing,

		"-dbpath": complete.PredictFiles("*"),

		"-apiaddress":  complete.PredictAnything,
		"-apiusername": complete.PredictAnything,
		"-apipassword": complete.PredictAnything,
		"-apitlscert":  complete.PredictFiles("*.cert"),
		"-apitlskey":   complete.PredictFiles("*.key"),
		"-apimaxlimit": complete.PredictAnything,
		"-apitimeout":  complete.PredictAnything,

		"-s":               complete.PredictAnything,
		"-factomdtimeout":  complete.PredictAnything,
		"-factomduser":     complete.PredictAnything,
		"-factomdpassword": complete.PredictAnything,
		//"-factomdcert":     complete.PredictFiles("*"),
		//"-factomdtls":      complete.PredictNothing,

		"-w":              complete.PredictAnything,
		"-wallettimeout":  complete.PredictAnything,
		"-walletuser":     complete.PredictAnything,
		"-walletpassword": complete.PredictAnything,
		//"-walletcert":     complete.PredictFiles("*"),
		//"-wallettls":      complete.PredictNothing,

		"-username": complete.PredictNothing,
		"-password": complete.PredictNothing,
		"-cert":     complete.PredictFiles("*"),
		"-certkey":  complete.PredictFiles("*"),
		"-tls":      complete.PredictNothing,

		"-y":                   complete.PredictNothing,
		"-installcompletion":   complete.PredictNothing,
		"-uninstallcompletion": complete.PredictNothing,

		"-ecadr": predictAddress(false, 1, "-ecadr", ""),

		"-whitelist":       complete.PredictAnything,
		"-blacklist":       complete.PredictAnything,
		"-ignorenewchains": complete.PredictNothing,

		"-networkid": complete.PredictSet("mainnet", "testnet", "localnet", "0x"),

		"-skipdbvalidation": complete.PredictNothing,
	}

	startScanHeight    uint64      // We parse the flag as unsigned.
	StartScanHeight    int32  = -1 // We work with the signed value.
	FactomScanInterval time.Duration
	LogDebug           bool
	DisablePending     bool
	FactomScanRetries  int64 = -1

	EsAdr factom.EsAddress
	ECAdr factom.ECAddress

	DBPath string

	APIAddress  string
	APIMaxLimit uint64
	APITimeout  time.Duration

	FactomClient = factom.NewClient()
	NetworkID    factom.NetworkID

	flagset    map[string]bool
	log        *logrus.Entry
	Completion *complete.Complete

	Whitelist, Blacklist Bytes32List
	ignoreNewChains      bool
	SkipDBValidation     bool

	HasAuth  bool
	Username string
	Password string

	HasTLS      bool
	TLSCertFile string
	TLSKeyFile  string
)

func init() {
	flagVar(&startScanHeight, "startscanheight")
	flagVar(&FactomScanRetries, "factomscanretries")
	flagVar(&FactomScanInterval, "factomscaninterval")
	flagVar(&LogDebug, "debug")
	flagVar(&DisablePending, "disablepending")

	flagVar(&DBPath, "dbpath")

	flagVar(&APIAddress, "apiaddress")
	flagVar(&APIMaxLimit, "apimaxlimit")
	flagVar(&APITimeout, "apitimeout")
	// Added in FatD authentication info.
	flagVar(&Username, "apiusername")
	flagVar(&Password, "apipassword")
	flagVar(&TLSCertFile, "apitlscert")
	flagVar(&TLSKeyFile, "apitlskey")

	flagVar(&ECAdr, "ecadr")
	flagVar(&EsAdr, "esadr")

	flagVar(&FactomClient.FactomdServer, "s")
	flagVar(&FactomClient.Factomd.Timeout, "factomdtimeout")
	flagVar(&FactomClient.Factomd.User, "factomduser")
	flagVar(&FactomClient.Factomd.Password, "factomdpassword")
	flagVar(&NetworkID, "networkid")
	//flagVar(&FactomClient.Factomd.TLSCertFile, "factomdcert")
	//flagVar(&FactomClient.Factomd.TLSEnable, "factomdtls")

	flagVar(&FactomClient.WalletdServer, "w")
	flagVar(&FactomClient.Walletd.Timeout, "wallettimeout")
	flagVar(&FactomClient.Walletd.User, "walletuser")
	flagVar(&FactomClient.Walletd.Password, "walletpassword")
	//flagVar(&FactomClient.Walletd.TLSCertFile, "walletcert")
	//flagVar(&FactomClient.Walletd.TLSEnable, "wallettls")

	flagVar(&Whitelist, "whitelist")
	flagVar(&Blacklist, "blacklist")
	flagVar(&ignoreNewChains, "ignorenewchains")
	flagVar(&SkipDBValidation, "skipdbvalidation")

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
	loadFromEnv(&FactomScanRetries, "factomscanretries")
	loadFromEnv(&FactomScanInterval, "factomscaninterval")
	loadFromEnv(&LogDebug, "debug")
	loadFromEnv(&DisablePending, "disablepending")

	loadFromEnv(&DBPath, "dbpath")

	loadFromEnv(&APIAddress, "apiaddress")

	loadFromEnv(&FactomClient.FactomdServer, "s")
	loadFromEnv(&FactomClient.Factomd.Timeout, "factomdtimeout")
	loadFromEnv(&FactomClient.Factomd.User, "factomduser")
	loadFromEnv(&FactomClient.Factomd.Password, "factomdpassword")
	//loadFromEnv(&FactomClient.Factomd.TLSCertFile, "factomdcert")
	//loadFromEnv(&FactomClient.Factomd.TLSEnable, "factomdtls")

	loadFromEnv(&FactomClient.WalletdServer, "w")
	loadFromEnv(&FactomClient.Walletd.Timeout, "walletdtimeout")
	loadFromEnv(&FactomClient.Walletd.User, "walletuser")
	loadFromEnv(&FactomClient.Walletd.Password, "walletpassword")
	//loadFromEnv(&FactomClient.Walletd.TLSCertFile, "walletcert")
	//loadFromEnv(&FactomClient.Walletd.TLSEnable, "wallettls")

	loadFromEnv(&ECAdr, "ecadr")
	loadFromEnv(&EsAdr, "esadr")

	if flagset["startscanheight"] {
		StartScanHeight = int32(startScanHeight)
	}
	if !flagset["networkid"] {
		NetworkID = factom.MainnetID()
	}
}

func Validate() {
	// Redact private data from debug output.
	factomdPassword := `""`
	if len(FactomClient.Factomd.Password) > 0 {
		factomdPassword = "<redacted>"
	}
	walletdPassword := `""`
	if len(FactomClient.Walletd.Password) > 0 {
		walletdPassword = "<redacted>"
	}
	apiPassword := `""`
	if len(Password) > 0 {
		apiPassword = "<redacted>"
	}

	log.Debugf("-dbpath            %#v", DBPath)
	log.Debugf("-apiaddress        %#v", APIAddress)
	debugPrintln()

	log.Debugf("-startscanheight   %v ", StartScanHeight)
	log.Debugf("-factomscanretries %v ", FactomScanRetries)
	log.Debugf("-factomscaninterval %v ", FactomScanInterval)
	debugPrintln()

	log.Debugf("-networkid      %v", NetworkID)
	log.Debugf("-s              %q", FactomClient.FactomdServer)
	log.Debugf("-factomdtimeout %v ", FactomClient.Factomd.Timeout)
	log.Debugf("-factomduser    %q", FactomClient.Factomd.User)
	log.Debugf("-factomdpass    %v ", factomdPassword)
	debugPrintln()

	log.Debugf("-w              %#v", FactomClient.WalletdServer)
	log.Debugf("-wallettimeout %v ", FactomClient.Walletd.Timeout)
	log.Debugf("-walletuser    %#v", FactomClient.Walletd.User)
	log.Debugf("-walletpass    %v ", walletdPassword)
	debugPrintln()

	log.Debugf("-apiusername    %#v", Username)
	log.Debugf("-apipassword    %v ", apiPassword)
	log.Debugf("-apitlscert     %#v", TLSCertFile)
	log.Debugf("-apitlskey      %#v", TLSKeyFile)
	debugPrintln()

	var err error
	DBPath, err = filepath.Abs(DBPath)
	if err != nil {
		log.Fatalf("-dbpath %v: %v", DBPath, err)
	}
	DBPath += fmt.Sprintf("%c", filepath.Separator)

	if factom.Bytes32(EsAdr).IsZero() {
		EsAdr, _ = ECAdr.GetEsAddress(context.TODO(), FactomClient)
	} else {
		ECAdr = EsAdr.ECAddress()
	}

	if IgnoreNewChains() && flagset["startscanheight"] {
		log.Fatal(
			"-startscanheight incompatible with -ignorenewchains and -whitelist")
	}

	if len(Username) > 0 || len(Password) > 0 {
		if len(Username) == 0 || len(Password) == 0 {
			log.Fatal("-apiusername and -apipassword must be used together")
		}
		HasAuth = true
	}
	if len(TLSCertFile) > 0 || len(TLSKeyFile) > 0 {
		if len(TLSCertFile) == 0 || len(TLSKeyFile) == 0 {
			log.Fatal("-apitlscert and -apitlskey must be used together")
		}
		HasTLS = true
	}
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
		case *int64:
			val, err := strconv.ParseInt(eVar, 10, 64)
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

func HasWhitelist() bool {
	return flagset["whitelist"]
}

func IgnoreNewChains() bool {
	return ignoreNewChains || HasWhitelist()
}
