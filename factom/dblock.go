package factom

type DBlock struct {
	Height  int64    `json:"-"`
	EBlocks []EBlock `json:"dbentries"`
}

type EBlock struct {
	ChainID      Bytes32 `json:"chainid"`
	KeyMR        Bytes32 `json:"keymr"`
	Entries      []Entry `json:"entrylist"`
	EBlockHeader `json:"header"`
}

type EBlockHeader struct {
	Height    int64   `json:"dbheight"`
	PrevKeyMR Bytes32 `json:"prevkeymr"`
}

type Entry struct {
	Hash      Bytes32 `json:"entryhash"`
	Timestamp Time    `json:"timestamp"`

	ChainID Bytes32 `json:"chainid"`
	Content Bytes   `json:"content"`
	ExtIDs  []Bytes `json:"extids"`
}

func DBlockByHeight(height int64) (*DBlock, error) {
	params := map[string]int64{"height": height}
	result := &struct {
		*DBlock `json:"dblock"`
	}{DBlock: &DBlock{}}
	if err := request("dblock-by-height", params, result); err != nil {
		return nil, err
	}
	return result.DBlock, nil
}

func GetEntryBlock(eblock *EBlock) error {
	params := map[string]*Bytes32{"keymr": &eblock.KeyMR}
	if err := request("entry-block", params, eblock); err != nil {
		return err
	}
	return nil
}

func GetEntry(entry *Entry) error {
	params := map[string]*Bytes32{"hash": &entry.Hash}
	if err := request("entry", params, entry); err != nil {
		return err
	}
	return nil
}

func (eb EBlock) IsNewChain() bool {
	return eb.PrevKeyMR == ZeroBytes32
}

var ZeroBytes32 Bytes32
