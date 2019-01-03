package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Factom-Asset-Tokens/base58"
	"github.com/Factom-Asset-Tokens/fatd/factom"
	"github.com/Factom-Asset-Tokens/fatd/fat0"
	"github.com/FactomProject/ed25519"
	"github.com/sirupsen/logrus"
)

// Environment variable name prefix
const envNamePrefix = "FATCLI_"

var (
	envNames = map[string]string{
		"debug": "DEBUG",

		"apiaddress": "API_ADDRESS",

		"w":              "WALLETD_SERVER",
		"wallettimeout":  "WALLETD_TIMEOUT",
		"walletuser":     "WALLETD_USER",
		"walletpassword": "WALLETD_PASSWORD",
		"walletcert":     "WALLETD_TLS_CERT",
		"wallettls":      "WALLETD_TLS_ENABLE",

		"s":               "FACTOMD_SERVER",
		"factomdtimeout":  "FACTOMD_TIMEOUT",
		"factomduser":     "FACTOMD_USER",
		"factomdpassword": "FACTOMD_PASSWORD",
		"factomdcert":     "FACTOMD_TLS_CERT",
		"factomdtls":      "FACTOMD_TLS_ENABLE",

		"ecpub": "ECPUB",
	}
	defaults = map[string]interface{}{
		"debug": false,

		"apiaddress": "http://localhost:8078",

		"w":              "localhost:8089",
		"wallettimeout":  time.Duration(0),
		"walletuser":     "",
		"walletpassword": "",
		"walletcert":     "",
		"wallettls":      false,

		"s":               "localhost:8088",
		"factomdtimeout":  time.Duration(0),
		"factomduser":     "",
		"factomdpassword": "",
		"factomdcert":     "",
		"factomdtls":      false,

		"chainid":  "Token Chain ID",
		"tokenid":  "",
		"identity": "Issuer Identity Chain used in Token Chain ID derivation",
		"ecpub":    "",

		"type":   "FAT-0",
		"supply": int64(0),
		"symbol": "",
		"name":   "",

		"coinbase": uint64(0),
	}
	descriptions = map[string]string{
		"debug": "Log debug messages",

		"apiaddress": "IPAddr:port# to bind to for serving the JSON RPC 2.0 API",

		"w":              "IPAddr:port# of factom-walletd API to use to access blockchain",
		"wallettimeout":  "Timeout for factom-walletd API requests, 0 means never timeout",
		"walletuser":     "Username for API connections to factom-walletd",
		"walletpassword": "Password for API connections to factom-walletd",
		"walletcert":     "The TLS certificate that will be provided by the factom-walletd API server",
		"wallettls":      "Set to true to use TLS when accessing the factom-walletd API",

		"s":               "IPAddr:port# of factomd API to use to access wallet",
		"factomdtimeout":  "Timeout for factomd API requests, 0 means never timeout",
		"factomduser":     "Username for API connections to factomd",
		"factomdpassword": "Password for API connections to factomd",
		"factomdcert":     "The TLS certificate that will be provided by the factomd API server",
		"factomdtls":      "Set to true to use TLS when accessing the factomd API",

		"chainid":  "Token Chain ID",
		"tokenid":  "Token ID used in Token Chain ID derivation",
		"identity": "Issuer Identity Chain ID used in Token Chain ID derivation",
		"ecpub":    "Entry Credit Public Address to use to pay for Factom entries",

		"sk1":    "Issuer's SK1 key as defined by their Identity Chain.",
		"type":   `FAT Token Type (e.g. "FAT-0")`,
		"supply": "Total number of issuable tokens. Must be a positive integer or -1 for unlimited.",
		"symbol": "Ticker symbol for the token (optional)",
		"name":   "Complete descriptive name of the token (optional)",

		"coinbase": "Create a coinbase transaction with the given amount. Requires -sk1.",
		"input":    "Add an -input ADDRESS:AMOUNT to the transaction. Can be specified multiple times.",
		"output":   "Add an -output ADDRESS:AMOUNT to the transaction. Can be specified multiple times.",
	}

	issuance = func() fat0.Issuance {
		i := fat0.Issuance{}
		i.ChainID = factom.NewBytes32(nil)
		return i
	}()
	chainID     = issuance.ChainID
	transaction = func() fat0.Transaction {
		tx := fat0.Transaction{
			Inputs:  fat0.AddressAmountMap{},
			Outputs: fat0.AddressAmountMap{},
		}
		tx.ChainID = chainID
		return tx
	}()
	coinbaseAmount uint64
	identity       = fat0.Identity{ChainID: factom.NewBytes32(nil)}
	sk1            = factom.Address{PrivateKey: new([ed25519.PrivateKeySize]byte)}
	address        = factom.Address{}
	ECPub          string
	metadata       string
	tokenID        string

	txHash *factom.Bytes32

	cmd string

	globalFlagSet = flag.NewFlagSet("fat-cli", flag.ContinueOnError)

	issueFlagSet    = flag.NewFlagSet("issue", flag.ExitOnError)
	transactFlagSet = flag.NewFlagSet("transact", flag.ExitOnError)

	LogDebug bool

	APIAddress string

	rpc = factom.RpcConfig

	flagIsSet = map[string]bool{}
	log       *logrus.Entry
)

func init() {
	flagVar(globalFlagSet, &LogDebug, "debug")

	flagVar(globalFlagSet, &APIAddress, "apiaddress")

	flagVar(globalFlagSet, &rpc.WalletServer, "w")
	flagVar(globalFlagSet, &rpc.WalletTimeout, "wallettimeout")
	flagVar(globalFlagSet, &rpc.WalletRPCUser, "walletuser")
	flagVar(globalFlagSet, &rpc.WalletRPCPassword, "walletpassword")
	flagVar(globalFlagSet, &rpc.WalletTLSCertFile, "walletcert")
	flagVar(globalFlagSet, &rpc.WalletTLSEnable, "wallettls")

	flagVar(globalFlagSet, &rpc.FactomdServer, "s")
	flagVar(globalFlagSet, &rpc.FactomdTimeout, "factomdtimeout")
	flagVar(globalFlagSet, &rpc.FactomdRPCUser, "factomduser")
	flagVar(globalFlagSet, &rpc.FactomdRPCPassword, "factomdpassword")
	flagVar(globalFlagSet, &rpc.FactomdTLSCertFile, "factomdcert")
	flagVar(globalFlagSet, &rpc.FactomdTLSEnable, "factomdtls")

	flagVar(globalFlagSet, &tokenID, "tokenid")
	flagVar(globalFlagSet, (*flagBytes32)(identity.ChainID), "identity")
	flagVar(globalFlagSet, (*flagBytes32)(chainID), "chainid")

	flagVar(issueFlagSet, (*ecpub)(&ECPub), "ecpub")
	flagVar(issueFlagSet, (*SecretKey)(sk1.PrivateKey), "sk1")
	flagVar(issueFlagSet, &issuance.Type, "type")
	flagVar(issueFlagSet, &issuance.Supply, "supply")
	flagVar(issueFlagSet, &issuance.Symbol, "symbol")
	flagVar(issueFlagSet, &issuance.Name, "name")

	flagVar(transactFlagSet, (*ecpub)(&ECPub), "ecpub")
	flagVar(transactFlagSet, (*SecretKey)(sk1.PrivateKey), "sk1")
	flagVar(transactFlagSet, &coinbaseAmount, "coinbase")
	flagVar(transactFlagSet, (addressAmountMap)(transaction.Inputs), "input")
	flagVar(transactFlagSet, (addressAmountMap)(transaction.Outputs), "output")

	// Add flags for self installing the CLI completion tool
	Completion.CLI.InstallName = "installcompletion"
	Completion.CLI.UninstallName = "uninstallcompletion"
	Completion.AddFlags(globalFlagSet)
}
func setFlagIsSet(f *flag.Flag) { flagIsSet[f.Name] = true }

func Parse() {
	args := os.Args[1:]
	if len(args) == 0 {
		return
	}
	setupLogger()
	globalFlagSet.Parse(args)
	args = globalFlagSet.Args()
	globalFlagSet.Visit(setFlagIsSet)
	if len(args) > 0 {
		cmd = args[0]
		args = args[1:]
	}

	var flagSet *flag.FlagSet
	switch cmd {
	case "issue":
		flagSet = issueFlagSet
	case "transact":
		flagSet = transactFlagSet
	case "balance":
		if len(args) == 1 {
			if err := address.UnmarshalJSON(
				[]byte(fmt.Sprintf("%#v", args[0]))); err != nil {
				return
			}
		}
	case "gettransaction":
		if len(args) == 1 {
			txHash = factom.NewBytes32(nil)
			if err := txHash.UnmarshalJSON(
				[]byte(fmt.Sprintf("%#v", args[0]))); err != nil {
				txHash = nil
				return
			}
		}
	default:
	}
	if flagSet != nil {
		flagSet.Parse(args)
		flagSet.Visit(setFlagIsSet)
	}

	// Load options from environment variables if they haven't been
	// specified on the command line.
	loadFromEnv(&LogDebug, "debug")

	loadFromEnv(&APIAddress, "apiaddress")

	loadFromEnv(&rpc.WalletServer, "w")
	loadFromEnv(&rpc.WalletTimeout, "walletdtimeout")
	loadFromEnv(&rpc.WalletRPCUser, "factomduser")
	loadFromEnv(&rpc.WalletRPCPassword, "factomdpassword")
	loadFromEnv(&rpc.WalletTLSCertFile, "factomdcert")
	loadFromEnv(&rpc.WalletTLSEnable, "factomdtls")

	loadFromEnv(&rpc.FactomdServer, "s")
	loadFromEnv(&rpc.FactomdTimeout, "factomdtimeout")
	loadFromEnv(&rpc.FactomdRPCUser, "factomduser")
	loadFromEnv(&rpc.FactomdRPCPassword, "factomdpassword")
	loadFromEnv(&rpc.FactomdTLSCertFile, "factomdcert")
	loadFromEnv(&rpc.FactomdTLSEnable, "factomdtls")
}

func Validate() error {
	if len(cmd) == 0 {
		return nil
	}
	// Redact private data from debug output.
	factomdRPCPassword := "\"\""
	if len(rpc.FactomdRPCPassword) > 0 {
		factomdRPCPassword = "<redacted>"
	}
	walletRPCPassword := "\"\""
	if len(rpc.WalletRPCPassword) > 0 {
		walletRPCPassword = "<redacted>"
	}

	log.Debugf("-apiaddress      %#v", APIAddress)
	debugPrintln()

	log.Debugf("-w             %#v", rpc.WalletServer)
	log.Debugf("-walletuser    %#v", rpc.WalletRPCUser)
	log.Debugf("-walletpass    %v ", walletRPCPassword)
	log.Debugf("-walletcert    %#v", rpc.WalletTLSCertFile)
	log.Debugf("-wallettimeout %v ", rpc.WalletTimeout)
	debugPrintln()

	log.Debugf("-s              %#v", rpc.FactomdServer)
	log.Debugf("-factomduser    %#v", rpc.FactomdRPCUser)
	log.Debugf("-factomdpass    %v ", factomdRPCPassword)
	log.Debugf("-factomdcert    %#v", rpc.FactomdTLSCertFile)
	log.Debugf("-factomdtimeout %v ", rpc.FactomdTimeout)
	debugPrintln()

	// Validate options
	switch cmd {
	case "issue":
		requireTokenChain()
		if err := requireFlags("sk1", "type", "supply", "ecpub"); err != nil {
			return err
		}
		if err := issuance.ValidData(); err != nil {
			return err
		}
	case "balance":
		requireTokenChain()
		zero := factom.Address{}
		if address.RCDHash() == zero.RCDHash() {
			return fmt.Errorf("no address specified")
		}
	case "transact":
		requireTokenChain()
		required := []string{"output"}
		if flagIsSet["coinbase"] || flagIsSet["sk1"] {
			if flagIsSet["input"] {
				return fmt.Errorf(
					"cannot specify -input with -coinbase and -sk1")
			}
			required = append(required, "coinbase", "sk1")
			if coinbaseAmount == 0 {
				return fmt.Errorf("-coinbase amount may not be zero")
			}
			a := factom.Address{}
			transaction.Inputs[*a.RCDHash()] = coinbaseAmount
		} else {
			required = append(required, "input")
		}
		if err := requireFlags(required...); err != nil {
			return err
		}
	case "gettransaction":
		requireTokenChain()
		if txHash == nil {
			return fmt.Errorf("no transaction entry hash specified")
		}
	case "stats":
		fallthrough
	case "getissuance":
		requireTokenChain()
	case "listtokens":
	case "help":
	case "":
		return fmt.Errorf("No command supplied")
	default:
		return fmt.Errorf("Invalid command: %v", cmd)
	}
	return nil
}

func requireTokenChain() error {
	if !flagIsSet["chainid"] {
		if !flagIsSet["tokenid"] || !flagIsSet["identity"] {
			return fmt.Errorf(
				"You must specify -chainid OR -tokenid AND -identity")
		}
		chainID := fat0.ChainID(tokenID, identity.ChainID)
		copy(issuance.ChainID[:], chainID[:])
	} else {
		if flagIsSet["tokenid"] || flagIsSet["identity"] {
			return fmt.Errorf(
				"You may not specify -chainid with -tokenid and -identity")
		}
	}
	return nil
}

func flagVar(f *flag.FlagSet, v interface{}, name string) {
	dflt := defaults[name]
	desc := description(name)
	switch v := v.(type) {
	case *string:
		f.StringVar(v, name, dflt.(string), desc)
	case *time.Duration:
		f.DurationVar(v, name, dflt.(time.Duration), desc)
	case *uint64:
		f.Uint64Var(v, name, dflt.(uint64), desc)
	case *int64:
		f.Int64Var(v, name, dflt.(int64), desc)
	case *bool:
		f.BoolVar(v, name, dflt.(bool), desc)
	case flag.Value:
		f.Var(v, name, desc)
	}
}

func loadFromEnv(v interface{}, flagName string) {
	if flagIsSet[flagName] {
		return
	}
	eName := envName(flagName)
	eVar, ok := os.LookupEnv(eName)
	if len(eVar) > 0 {
		switch v := v.(type) {
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
	if _, ok := envNames[flagName]; ok {
		return fmt.Sprintf("%s\nEnvironment variable: %v",
			descriptions[flagName], envName(flagName))
	}
	return descriptions[flagName]
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

type flagBytes32 factom.Bytes32

// String returns the hex encoded data of b.
func (b *flagBytes32) String() string {
	if b == nil {
		return ""
	}
	return (*factom.Bytes32)(b).String()
}
func (b *flagBytes32) Set(data string) error {
	return (*factom.Bytes32)(b).UnmarshalJSON([]byte(fmt.Sprintf("%#v", data)))
}

type SecretKey [ed25519.PrivateKeySize]byte

// String returns the hex encoded data of b.
func (sk *SecretKey) String() string {
	if sk == nil {
		return ""
	}
	return "<redacted>"
}
func (sk *SecretKey) Set(data string) error {
	if len(data) != 53 {
		return fmt.Errorf("invalid length")
	}
	if data[0:3] != "sk1" {
		return fmt.Errorf("invalid prefix")
	}
	b, _, err := base58.CheckDecode(data, 3)
	if err != nil {
		return err
	}
	copy(sk[:], b)
	ed25519.GetPublicKey((*[ed25519.PrivateKeySize]byte)(sk))
	return nil
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

func requireFlags(names ...string) error {
	missing := []string{}
	for _, n := range names {
		if !flagIsSet[n] {
			missing = append(missing, "-"+n)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required flags: %v", missing)
	}
	return nil
}

type addressAmountMap fat0.AddressAmountMap

func (a addressAmountMap) Set(data string) error {
	s := strings.SplitN(data, ":", 2)
	if len(s) != 2 {
		return fmt.Errorf("invalid format")
	}
	adr := factom.Address{}
	if s[0] != "coinbase" {
		if err := adr.UnmarshalJSON(
			[]byte(fmt.Sprintf("%#v", s[0]))); err != nil {
			return fmt.Errorf("invalid address: %v", err)
		}
	}
	if _, ok := a[*adr.RCDHash()]; ok {
		return fmt.Errorf("duplicate address: %v", adr.RCDHash())
	}
	amount, err := strconv.ParseUint(s[1], 10, 64)
	if err != nil {
		return fmt.Errorf("invalid amount: %v", err)
	}
	if amount == 0 {
		return fmt.Errorf("invalid amount: may not be zero")
	}
	a[*adr.RCDHash()] = amount
	return nil
}
func (a addressAmountMap) String() string {
	return fmt.Sprintf("%v", fat0.AddressAmountMap(a))
}
