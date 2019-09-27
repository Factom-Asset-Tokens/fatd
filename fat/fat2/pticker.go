package fat2

import "fmt"

// PPTicker is an internal representation of a PegNet asset type
type PTicker int

const (
	PTickerInvalid PTicker = iota
	PTickerPEG
	PTickerUSD
	PTickerEUR
	PTickerJPY
	PTickerGBP
	PTickerCAD
	PTickerCHF
	PTickerINR
	PTickerSGD
	PTickerCNY
	PTickerHKD
	PTickerKRW
	PTickerBRL
	PTickerPHP
	PTickerMXN
	PTickerXAU
	PTickerXAG
	PTickerXBT
	PTickerETH
	PTickerLTC
	PTickerRVN
	PTickerXBC
	PTickerFCT
	PTickerBNB
	PTickerXLM
	PTickerADA
	PTickerXMR
	PTickerDAS
	PTickerZEC
	PTickerDCR
	PTickerMax
)

// TODO: figure out how to handle "tUSD" and other testnet types
//       or convince PegNet core devs that the differentiation is unnecessary
//       due to the existence of multiple testnets
var validPTickerStrings = []string{
	"PEG",
	"pUSD",
	"pEUR",
	"pJPY",
	"pGBP",
	"pCAD",
	"pCHF",
	"pINR",
	"pSGD",
	"pCNY",
	"pHKD",
	"pKRW",
	"pBRL",
	"pPHP",
	"pMXN",
	"pXAU",
	"pXAG",
	"pXBT",
	"pETH",
	"pLTC",
	"pRVN",
	"pXBC",
	"pFCT",
	"pBNB",
	"pXLM",
	"pADA",
	"pXMR",
	"pDAS",
	"pZEC",
	"pDCR",
}

var validPTickers = func() map[string]PTicker {
	pTickers := make(map[string]PTicker, len(validPTickerStrings))
	for i, str := range validPTickerStrings {
		pTickers[str] = PTicker(i + 1)
	}
	return pTickers
}()

// UnmarshalJSON unmarshals the bytes into a PTicker and returns an error
// if the ticker is invalid
func (t *PTicker) UnmarshalJSON(data []byte) error {
	ticker := string(data)
	// When unmarshalling, the bytes passed in are []byte("\"PEG\"") rather
	// than just[]byte("PEG") so we must ensure that we take the quotes into
	// account here
	if len(ticker) < 3 || ticker[0] != "\""[0] || ticker[len(ticker)-1] != "\""[0] {
		return fmt.Errorf("invalid token type")
	}
	pTicker, ok := validPTickers[ticker[1:len(ticker)-1]]
	if !ok {
		*t = PTickerInvalid
		return fmt.Errorf("invalid token type")
	}
	*t = pTicker
	return nil
}

// MarshalJSON marshals the PTicker into the bytes that represent it in JSON
func (t PTicker) MarshalJSON() ([]byte, error) {
	if t <= PTickerInvalid || PTickerMax <= t {
		return nil, fmt.Errorf("invalid token type")
	}
	pTickerString := validPTickerStrings[int(t)]
	// Ensure that quotes are included in the resulting bytes as it
	// will be used in the JSON exactly as marshalled
	return []byte(fmt.Sprintf("\"%s\"", pTickerString)), nil
}
