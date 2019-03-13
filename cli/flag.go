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
	"github.com/Factom-Asset-Tokens/fatd/fat"
	"github.com/Factom-Asset-Tokens/fatd/fat/fat0"
	"github.com/Factom-Asset-Tokens/fatd/fat/fat1"
	"github.com/posener/complete"
	"github.com/sirupsen/logrus"
)

// Environment variable name prefix
const envNamePrefix = "FATCLI_"

var (
	flagMap = map[string]struct {
		SubCommand  string
		EnvName     string
		Default     interface{}
		Var         map[string]interface{}
		Description string
		Predictor   complete.Predictor
		IsSet       bool
	}{"debug": {
		EnvName:     "DEBUG",
		Description: "Log debug messages",
		Predictor:   complete.PredictNothing,
		Var:         map[string]interface{}{"global": &LogDebug},
	}, "apiaddress": {
		EnvName:     "API_ADDRESS",
		Default:     "http://localhost:8078",
		Description: "IPAddr:port# to bind to for serving the JSON RPC 2.0 API",
		Predictor:   complete.PredictAnything,
		Var:         map[string]interface{}{"global": &APIAddress},
	}, "w": {
		EnvName:     "WALLETD_SERVER",
		Default:     "localhost:8089",
		Description: "IPAddr:port# of factom-walletd API to use to access blockchain",
		Predictor:   complete.PredictAnything,
		Var:         map[string]interface{}{"global": &rpc.WalletServer},
	}, "wallettimeout": {
		EnvName:     "WALLETD_TIMEOUT",
		Description: "Timeout for factom-walletd API requests, 0 means never timeout",
		Predictor:   complete.PredictAnything,
		Var:         map[string]interface{}{"global": &rpc.WalletTimeout},
	}, "walletuser": {
		EnvName:     "WALLETD_USER",
		Description: "Username for API connections to factom-walletd",
		Predictor:   complete.PredictAnything,
		Var:         map[string]interface{}{"global": &rpc.WalletRPCUser},
	}, "walletpassword": {
		EnvName:     "WALLETD_PASSWORD",
		Description: "Password for API connections to factom-walletd",
		Predictor:   complete.PredictAnything,
		Var:         map[string]interface{}{"global": &rpc.WalletRPCPassword},
	}, "walletcert": {
		EnvName:     "WALLETD_TLS_CERT",
		Description: "The TLS certificate that will be provided by the factom-walletd API server",
		Predictor:   complete.PredictFiles("*"),
		Var:         map[string]interface{}{"global": &rpc.WalletTLSCertFile},
	}, "wallettls": {
		EnvName:     "WALLETD_TLS_ENABLE",
		Description: "Set to true to use TLS when accessing the factom-walletd API",
		Predictor:   complete.PredictNothing,
		Var:         map[string]interface{}{"global": &rpc.WalletTLSEnable},
	}, "s": {
		EnvName:     "FACTOMD_SERVER",
		Default:     "localhost:8088",
		Description: "IPAddr:port# of factomd API to use to access wallet",
		Predictor:   complete.PredictAnything,
		Var:         map[string]interface{}{"global": &rpc.FactomdServer},
	}, "factomdtimeout": {
		EnvName:     "FACTOMD_TIMEOUT",
		Description: "Timeout for factomd API requests, 0 means never timeout",
		Predictor:   complete.PredictAnything,
		Var:         map[string]interface{}{"global": &rpc.FactomdTimeout},
	}, "factomduser": {
		EnvName:     "FACTOMD_USER",
		Description: "Username for API connections to factomd",
		Predictor:   complete.PredictAnything,
		Var:         map[string]interface{}{"global": &rpc.FactomdRPCUser},
	}, "factomdpassword": {
		EnvName:     "FACTOMD_PASSWORD",
		Description: "Password for API connections to factomd",
		Predictor:   complete.PredictAnything,
		Var:         map[string]interface{}{"global": &rpc.FactomdRPCPassword},
	}, "factomdcert": {
		EnvName:     "FACTOMD_TLS_CERT",
		Description: "The TLS certificate that will be provided by the factomd API server",
		Predictor:   complete.PredictFiles("*"),
		Var:         map[string]interface{}{"global": &rpc.FactomdTLSCertFile},
	}, "factomdtls": {
		EnvName:     "FACTOMD_TLS_ENABLE",
		Description: "Set to true to use TLS when accessing the factomd API",
		Predictor:   complete.PredictNothing,
		Var:         map[string]interface{}{"global": &rpc.FactomdTLSEnable},
	}, "ecpub": {
		SubCommand:  "issue|transactFAT0|transactFAT1",
		EnvName:     "ECPUB",
		Description: "Entry Credit Public Address to use to pay for Factom entries",
		Predictor:   predictAddress(false, 1, "-ecpub", ""),
		Var:         map[string]interface{}{"global": (*ECPub)(&ecpub)},
	}, "chainid": {
		Description: "Token Chain ID",
		Predictor:   complete.PredictAnything,
		Var:         map[string]interface{}{"global": (*flagBytes32)(chainID)},
	}, "tokenid": {
		Description: "Token ID used in Token Chain ID derivation",
		Predictor:   complete.PredictAnything,
		Var:         map[string]interface{}{"global": &tokenID},
	}, "identity": {
		Description: "Issuer Identity Chain ID used in Token Chain ID derivation",
		Predictor:   complete.PredictAnything,
		Var:         map[string]interface{}{"global": (*flagBytes32)(identity.ChainID)},
	}, "type": {
		SubCommand:  "issue",
		Description: `FAT Token Type (e.g. "FAT-0")`,
		Predictor:   complete.PredictSet("FAT-0", "FAT-1"),
		Var:         map[string]interface{}{"global": &issuance.Type},
	}, "sk1": {
		SubCommand:  "issue|transactFAT0|transactFAT1",
		Description: "Issuer's SK1 key as defined by their Identity Chain.",
		Predictor:   complete.PredictAnything,
		Var:         map[string]interface{}{"global": (*SecretKey)(sk1.PrivateKey())},
	}, "supply": {
		SubCommand:  "issue",
		Description: "Total number of issuable tokens. Must be a positive integer or -1 for unlimited.",
		Predictor:   complete.PredictAnything,
		Var:         map[string]interface{}{"global": &issuance.Supply},
	}, "symbol": {
		SubCommand:  "issue",
		Description: "Ticker symbol for the token (optional)",
		Predictor:   complete.PredictAnything,
		Var:         map[string]interface{}{"global": &issuance.Symbol},
	}, "coinbase": {
		SubCommand:  "transactFAT0|transactFAT1",
		Description: "Create a coinbase transaction with the given amount. Requires -sk1.",
		Predictor:   complete.PredictAnything,
		Var: map[string]interface{}{
			"transactFAT0": (*Amount)(&coinbaseAmount),
			"transactFAT1": (*NFTokens)(&coinbaseNFTokens)},
	}, "input": {
		SubCommand:  "transactFAT0|transactFAT1",
		Description: "Add an -input ADDRESS:AMOUNT to the transaction. Can be specified multiple times.",
		Predictor:   predictAddress(true, 1, "-input", ":"),
		Var: map[string]interface{}{
			"transactFAT0": (AddressAmountMap)(FAT0transaction.Inputs),
			"transactFAT1": (AddressNFTokensMap)(FAT1transaction.Inputs)},
	}, "output": {
		SubCommand:  "transactFAT0|transactFAT1",
		Description: "Add an -output ADDRESS:AMOUNT to the transaction. Can be specified multiple times.",
		Predictor:   predictAddress(true, 1, "-output", ":"),
		Var: map[string]interface{}{
			"transactFAT0": (AddressAmountMap)(FAT0transaction.Outputs),
			"transactFAT1": (AddressNFTokensMap)(FAT1transaction.Outputs)},
	}, "y": {
		Predictor: complete.PredictNothing,
	}, "installcompletion": {
		Predictor: complete.PredictNothing,
	}, "uninstallcompletion": {
		Predictor: complete.PredictNothing,
	}}

	Completion = func() *complete.Complete {
		cmd := complete.Command{Flags: complete.Flags{},
			Sub: complete.Commands{
				"transactFAT0": complete.Command{Flags: complete.Flags{},
					Args: complete.PredictAnything},
				"transactFAT1": complete.Command{Flags: complete.Flags{},
					Args: complete.PredictAnything},
				"issue": complete.Command{Flags: complete.Flags{},
					Args: complete.PredictAnything},
				"balance": complete.Command{
					Args: predictAddress(true, 1, "", "")},
				"version": complete.Command{Flags: complete.Flags{},
					Args: complete.PredictNothing},
			}}
		// Set sub command args
		for name, flag := range flagMap {
			// Set global flags
			if len(flag.SubCommand) == 0 {
				cmd.Flags["-"+name] = flag.Predictor
				continue
			}
			// Set sub command flags
			subCmds := strings.Split(flag.SubCommand, "|")
			for _, subCmdName := range subCmds {
				cmd.Sub[subCmdName].Flags["-"+name] = flag.Predictor
			}
		}

		cmplt := complete.New("fat-cli", cmd)
		// Add flags for self installing the CLI completion tool
		cmplt.CLI.InstallName = "installcompletion"
		cmplt.CLI.UninstallName = "uninstallcompletion"
		cmplt.AddFlags(globalFlagSet)
		return cmplt
	}()

	globalFlagSet       = flag.NewFlagSet("fat-cli", flag.ContinueOnError)
	issueFlagSet        = flag.NewFlagSet("issue", flag.ExitOnError)
	transactFAT0FlagSet = flag.NewFlagSet("transactFAT0", flag.ExitOnError)
	transactFAT1FlagSet = flag.NewFlagSet("transactFAT1", flag.ExitOnError)
	SubCommand          string

	// Global variables that hold flag vars
	chainID  = factom.NewBytes32(nil)
	issuance = func() fat.Issuance {
		i := fat.Issuance{}
		i.ChainID = chainID
		return i
	}()

	FAT0transaction = func() fat0.Transaction {
		tx := fat0.Transaction{
			Inputs:  fat0.AddressAmountMap{},
			Outputs: fat0.AddressAmountMap{},
		}
		tx.ChainID = chainID
		return tx
	}()
	coinbaseAmount uint64

	FAT1transaction = func() fat1.Transaction {
		tx := fat1.Transaction{
			Inputs:  fat1.AddressNFTokensMap{},
			Outputs: fat1.AddressNFTokensMap{},
		}
		tx.ChainID = chainID
		return tx
	}()
	coinbaseNFTokens = fat1.NFTokens{}

	identity = fat.Identity{ChainID: factom.NewBytes32(nil)}
	sk1      = factom.Address{}
	address  = factom.Address{}
	coinbase = factom.Address{}
	ecpub    string
	metadata string
	tokenID  string

	txHash *factom.Bytes32

	LogDebug bool

	APIAddress string

	rpc = factom.RpcConfig

	log *logrus.Entry
)

func setFlag(f *flag.Flag) {
	flag := flagMap[f.Name]
	flag.IsSet = true
	flagMap[f.Name] = flag
}

func init() {
	for name, flagS := range flagMap {
		// Set global flags
		if len(flagS.SubCommand) == 0 {
			flagVar(globalFlagSet, name, flagS.Var["global"])
			continue
		}
		// Set sub command flags
		subCmds := strings.Split(flagS.SubCommand, "|")
		var flagSet *flag.FlagSet
		for _, subCmdName := range subCmds {
			Var := flagS.Var["global"]
			if Var == nil {
				Var = flagS.Var[subCmdName]
			}
			switch subCmdName {
			case "issue":
				flagSet = issueFlagSet
			case "transactFAT0":
				flagSet = transactFAT0FlagSet
			case "transactFAT1":
				flagSet = transactFAT1FlagSet
			default:
				panic("invalid sub command: " + subCmdName)
			}
			flagVar(flagSet, name, Var)
		}
	}
}

func ParseCLI() {
	args := os.Args[1:]
	if len(args) == 0 {
		return
	}
	setupLogger()
	globalFlagSet.Parse(args)
	args = globalFlagSet.Args()
	globalFlagSet.Visit(setFlag)
	if len(args) > 0 {
		SubCommand = args[0]
		args = args[1:]
	}

	var flagSet *flag.FlagSet
	switch SubCommand {
	case "issue":
		flagSet = issueFlagSet
	case "transactFAT0":
		flagSet = transactFAT0FlagSet
	case "transactFAT1":
		flagSet = transactFAT1FlagSet
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
		flagSet.Visit(setFlag)
	}

	// Load options from environment variables if they haven't been
	// specified on the command line.
	for name, flag := range flagMap {
		if len(flag.EnvName) == 0 {
			continue
		}
		loadFromEnv(name, flag.Var["global"])
	}
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

func Validate() error {
	if len(SubCommand) == 0 {
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

	// Validate SubCommand
	switch SubCommand {
	// These SubCommands require further flag validation.
	case "issue", "balance", "transactFAT0", "transactFAT1",
		"gettransaction", "stats", "getissuance":
	// These SubCommands do not require any flags.
	case "listtokens", "version", "help":
		return nil

	case "":
		return fmt.Errorf("No command supplied")
	// Invalid SubCommands.
	default:
		return fmt.Errorf("Invalid command: %v", SubCommand)
	}

	if err := requireTokenChain(); err != nil {
		return err
	}

	switch SubCommand {
	case "issue":
		if err := requireFlags("sk1", "supply", "ecpub"); err != nil {
			return err
		}
		if err := issuance.ValidData(); err != nil {
			return err
		}
	case "balance":
		zero := factom.Address{}
		if address.RCDHash() == zero.RCDHash() {
			return fmt.Errorf("no address specified")
		}
	case "transactFAT0", "transactFAT1":
		required := []string{"output"}
		if flagMap["coinbase"].IsSet || flagMap["sk1"].IsSet {
			if flagMap["input"].IsSet {
				return fmt.Errorf(
					"cannot specify -input with -coinbase and -sk1")
			}
			required = append(required, "coinbase", "sk1")
			FAT0transaction.Inputs[*coinbase.RCDHash()] = coinbaseAmount
			FAT1transaction.Inputs[*coinbase.RCDHash()] = coinbaseNFTokens
		} else {
			required = append(required, "input")
		}
		if err := requireFlags(required...); err != nil {
			return err
		}
	case "gettransaction":
		if txHash == nil {
			return fmt.Errorf("no transaction entry hash specified")
		}
	}
	return nil
}
func debugPrintln() {
	if LogDebug {
		fmt.Println()
	}
}

func requireTokenChain() error {
	if !flagMap["chainid"].IsSet {
		if !flagMap["tokenid"].IsSet || !flagMap["identity"].IsSet {
			return fmt.Errorf(
				"You must specify -chainid OR -tokenid AND -identity")
		}
		nameIDs := fat.NameIDs(tokenID, identity.ChainID)
		if !fat.ValidTokenNameIDs(nameIDs) {
			return fmt.Errorf("The given -tokenid and -identity do not form a valid FAT Chain.")
		}
		chainID := fat.ChainID(tokenID, identity.ChainID)
		copy(issuance.ChainID[:], chainID[:])
	} else {
		if flagMap["tokenid"].IsSet || flagMap["identity"].IsSet {
			return fmt.Errorf(
				"You may not specify -chainid with -tokenid and -identity")
		}
	}
	return nil
}

func flagVar(f *flag.FlagSet, name string, val interface{}) {
	dflt := flagMap[name].Default
	desc := description(name)
	switch val := val.(type) {
	case *string:
		Default := ""
		if dflt != nil {
			Default = dflt.(string)
		}
		f.StringVar(val, name, Default, desc)
	case *time.Duration:
		Default := time.Duration(0)
		if dflt != nil {
			Default = dflt.(time.Duration)
		}
		f.DurationVar(val, name, Default, desc)
	case *uint64:
		Default := uint64(0)
		if dflt != nil {
			Default = dflt.(uint64)
		}
		f.Uint64Var(val, name, Default, desc)
	case *int64:
		Default := int64(0)
		if dflt != nil {
			Default = dflt.(int64)
		}
		f.Int64Var(val, name, Default, desc)
	case *bool:
		Default := false
		if dflt != nil {
			Default = dflt.(bool)
		}
		f.BoolVar(val, name, Default, desc)
	case flag.Value:
		f.Var(val, name, desc)
	}
}
func description(flagName string) string {
	if len(flagMap[flagName].EnvName) == 0 {
		return flagMap[flagName].Description
	}
	return fmt.Sprintf("%s\nEnvironment variable: %v",
		flagMap[flagName].Description, envName(flagName))
}
func envName(flagName string) string {
	return envNamePrefix + flagMap[flagName].EnvName
}

func loadFromEnv(flagName string, val interface{}) {
	if flagMap[flagName].IsSet {
		return
	}
	eName := envName(flagName)
	eVar, ok := os.LookupEnv(eName)
	if len(eVar) > 0 {
		switch val := val.(type) {
		case *string:
			*val = eVar
		case *time.Duration:
			duration, err := time.ParseDuration(eVar)
			if err != nil {
				log.Fatalf("Environment Variable %v: "+
					"time.ParseDuration(\"%v\"): %v",
					eName, eVar, err)
			}
			*val = duration
		case *uint64:
			v, err := strconv.ParseUint(eVar, 10, 64)
			if err != nil {
				log.Fatalf("Environment Variable %v: "+
					"strconv.ParseUint(\"%v\", 10, 64): %v",
					eName, eVar, err)
			}
			*val = v
		case *bool:
			if ok {
				*val = true
			}
		}
	}
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

type SecretKey factom.PrivateKey

func (sk SecretKey) String() string {
	return "<redacted>"
}
func (sk *SecretKey) Set(data string) error {
	if err := decodeBase58String(sk[:], data, 53, "sk1"); err != nil {
		return err
	}
	(*factom.PrivateKey)(sk).PublicKey()
	return nil
}
func decodeBase58String(dst []byte, data string,
	expectedLen int, prefix string) error {
	if len(data) != expectedLen {
		return fmt.Errorf("invalid length")
	}
	if string(data[0:len(prefix)]) != prefix {
		return fmt.Errorf("invalid prefix")
	}
	b, _, err := base58.CheckDecode(string(data), len(prefix))
	if err != nil {
		return err
	}
	copy(dst, b)
	return nil
}

type ECPub string

// String returns the hex encoded data of b.
func (ec ECPub) String() string {
	return string(ec)
}
func (ec *ECPub) Set(data string) error {
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
	*ec = ECPub(data)
	return nil
}

func requireFlags(names ...string) error {
	missing := []string{}
	for _, n := range names {
		if !flagMap[n].IsSet {
			missing = append(missing, "-"+n)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required flags: %v", missing)
	}
	return nil
}

type AddressAmountMap fat0.AddressAmountMap

func (m AddressAmountMap) Set(data string) error {
	s := strings.SplitN(data, ":", 2)
	if len(s) != 2 {
		return fmt.Errorf("invalid format")
	}
	adr := factom.RCDHash{}
	if s[0] == "coinbase" {
		adr = *coinbase.RCDHash()
	} else {
		if err := adr.FromString(s[0]); err != nil {
			return fmt.Errorf("invalid address: %v", err)
		}
	}
	if _, ok := m[adr]; ok {
		return fmt.Errorf("duplicate address: %v", adr)
	}
	var amount uint64
	if err := (*Amount)(&amount).Set(s[1]); err != nil {
		return err
	}
	m[adr] = amount
	return nil
}
func (m AddressAmountMap) String() string {
	return fmt.Sprintf("%v", fat0.AddressAmountMap(m))
}

type AddressNFTokensMap fat1.AddressNFTokensMap

func (m AddressNFTokensMap) Set(data string) error {
	s := strings.SplitN(data, ":", 2)
	if len(s) != 2 {
		return fmt.Errorf("invalid format")
	}
	adr := factom.RCDHash{}
	if s[0] == "coinbase" {
		adr = *coinbase.RCDHash()
	} else {
		if err := adr.FromString(s[0]); err != nil {
			return fmt.Errorf("invalid address: %v", err)
		}
	}
	if _, ok := m[adr]; ok {
		return fmt.Errorf("duplicate address: %v", adr)
	}

	tkns := make(fat1.NFTokens)
	if err := NFTokens(tkns).Set(s[1]); err != nil {
		return err
	}

	m[adr] = tkns
	return nil
}
func (m AddressNFTokensMap) String() string {
	return fmt.Sprintf("%v", fat1.AddressNFTokensMap(m))
}

type NFTokens fat1.NFTokens

func (tkns NFTokens) Set(data string) error {
	if data[0] != '[' || data[len(data)-1] != ']' {
		return fmt.Errorf("invalid NFTokens format")
	}
	tknStrs := strings.Split(data[1:len(data)-1], ",")
	for _, tknStr := range tknStrs {
		var tknIDs fat1.NFTokensSetter
		if strings.Contains(tknStr, "-") {
			tknRangeStrs := strings.Split(tknStr, "-")
			if len(tknRangeStrs) != 2 {
				return fmt.Errorf("invalid NFTokenIDRange format: %v", tknStr)
			}
			var err error
			var min, max uint64
			if min, err = strconv.ParseUint(tknRangeStrs[0], 10, 64); err != nil {
				return fmt.Errorf("invalid NFTokenIDRange.Min: %v", err)
			}
			if max, err = strconv.ParseUint(tknRangeStrs[1], 10, 64); err != nil {
				return fmt.Errorf("invalid NFTokenIDRange.Max: %v", err)
			}
			tknIDRange := fat1.NFTokenIDRange{
				Min: fat1.NFTokenID(min), Max: fat1.NFTokenID(max)}
			if err := tknIDRange.Valid(); err != nil {
				return fmt.Errorf("invalid NFTokenIDRange: %v", err)
			}
			tknIDs = tknIDRange
		} else {
			var tknID uint64
			var err error
			if tknID, err = strconv.ParseUint(tknStr, 10, 64); err != nil {
				return fmt.Errorf("invalid NFTokenID: %v", err)

			}
			tknIDs = fat1.NFTokenID(tknID)
		}
		if err := fat1.NFTokens(tkns).Set(tknIDs); err != nil {
			return fmt.Errorf("invalid NFTokens: %v", err)
		}
	}
	return nil
}
func (tkns NFTokens) String() string {
	return fmt.Sprintf("%v", fat1.NFTokens(tkns))
}

type Amount uint64

func (a *Amount) Set(data string) error {
	amount, err := strconv.ParseUint(data, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid amount: %v", err)
	}
	if amount == 0 {
		return fmt.Errorf("invalid amount: may not be zero")
	}
	*a = Amount(amount)
	return nil
}
func (a Amount) String() string {
	return fmt.Sprintf("%v", uint64(a))
}
