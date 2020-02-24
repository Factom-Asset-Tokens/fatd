package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/Factom-Asset-Tokens/factom"
	"github.com/Factom-Asset-Tokens/factom/fat"
	"github.com/Factom-Asset-Tokens/factom/fat104"
)

type BinaryFile struct {
	Path string
	Data []byte
}

func (f *BinaryFile) Set(path string) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	f.Path = path
	f.Data = data
	return nil
}
func (f BinaryFile) String() string {
	return f.Path
}
func (BinaryFile) Type() string {
	return "Binary File"
}

// JSONOrFile parses a flag as either raw JSON, or a path to a file that
// contains raw JSON.
type JSONOrFile json.RawMessage

func (js *JSONOrFile) Set(jsonOrPath string) error {
	if json.Valid([]byte(jsonOrPath)) {
		*js = JSONOrFile(jsonOrPath)
		return nil
	}
	data, err := ioutil.ReadFile(jsonOrPath)
	if err != nil {
		return fmt.Errorf("invalid JSON and %w", err)
	}
	if !json.Valid(data) {
		return fmt.Errorf("file %q does not contain valid JSON", jsonOrPath)
	}
	*js = JSONOrFile(data)
	return nil
}
func (js JSONOrFile) String() string {
	return string(js)
}
func (JSONOrFile) Type() string {
	return "JSON or file"
}

// JSON parses a flag as raw json.
type JSON json.RawMessage

func (r *JSON) Set(data string) error {
	if !json.Valid([]byte(data)) {
		return fmt.Errorf("invalid JSON")
	}
	*r = JSON(data)
	return nil
}

func (r JSON) String() string {
	return string(r)
}

func (JSON) Type() string {
	return "JSON"
}

type ABI fat104.ABI

func (abi *ABI) Set(text string) error {
	var js JSONOrFile
	if err := js.Set(text); err != nil {
		return err
	}
	return json.Unmarshal([]byte(js), abi)
}
func (abi ABI) String() string {
	s, _ := json.Marshal(fat104.ABI(abi))
	return string(s)
}
func (ABI) Type() string {
	return JSONOrFile{}.Type()
}

// FAAddressList parses a flag as a comma separated list of FAAddress. Multiple
// uses of the flag append to the list.
type FAAddressList []factom.FAAddress

func (adrs *FAAddressList) Set(adrStr string) error {
	adr, err := factom.NewFAAddress(adrStr)
	if err != nil {
		return err
	}
	*adrs = append(*adrs, adr)
	return nil
}
func (adrs FAAddressList) String() string {
	return fmt.Sprintf("%#v", adrs)
}
func (FAAddressList) Type() string {
	return "FAAddress"
}

// PaginationOrder parses a flag as ascending or descending.
type PaginationOrder string

func (o *PaginationOrder) Set(str string) error {
	str = strings.ToLower(str)
	switch str {
	case "asc", "ascending", "earliest":
		*o = "asc"
	case "des", "desc", "descending", "latest":
		*o = "desc"
	default:
		return fmt.Errorf(`must be "asc" or "desc"`)
	}
	return nil
}
func (o PaginationOrder) String() string {
	return string(o)
}
func (PaginationOrder) Type() string {
	return "asc|desc"
}

// ECEsAddress parses a flag as an EC or Es Address.
type ECEsAddress struct {
	EC factom.ECAddress
	Es factom.EsAddress
}

func (e *ECEsAddress) Set(adrStr string) error {
	if err := e.EC.Set(adrStr); err != nil {
		if err := e.Es.Set(adrStr); err != nil {
			return err
		}
		e.EC = e.Es.ECAddress()
	}
	return nil
}
func (e ECEsAddress) String() string {
	return e.EC.String()
}
func (ECEsAddress) Type() string {
	return "<EC | Es>"
}

// FATType parses a flag as a fat.Type like "FAT-0".
type FATType fat.Type

func (t *FATType) Set(typeStr string) error {
	typeStr = strings.ToUpper(typeStr)
	switch typeStr {
	case "FAT0":
		typeStr = "FAT-0"
	case "FAT1":
		typeStr = "FAT-1"
	}
	return (*fat.Type)(t).Set(typeStr)
}

func (t FATType) String() string {
	return fat.Type(t).String()
}
func (FATType) Type() string {
	return `<"FAT-0" | "FAT-1">`
}
