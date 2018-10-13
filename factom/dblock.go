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

	ChainID *Bytes32 `json:"-"`
	Content Bytes    `json:"content"`
	ExtIDs  []Bytes  `json:"extids"`
}

func (db *DBlock) Get() error {
	params := map[string]int64{"height": db.Height}
	result := &struct {
		*DBlock `json:"dblock"`
	}{DBlock: db}
	if err := request("dblock-by-height", params, result); err != nil {
		return err
	}
	return nil
}

func (eb *EBlock) Get() error {
	params := map[string]*Bytes32{"keymr": &eb.KeyMR}
	if err := request("entry-block", params, eb); err != nil {
		return err
	}
	for i, _ := range eb.Entries {
		eb.Entries[i].ChainID = &eb.ChainID
	}
	return nil
}

func (e *Entry) Get() error {
	params := map[string]*Bytes32{"hash": &e.Hash}
	if err := request("entry", params, e); err != nil {
		return err
	}
	return nil
}

func (eb EBlock) IsNewChain() bool {
	return eb.PrevKeyMR == ZeroBytes32
}

var ZeroBytes32 Bytes32
