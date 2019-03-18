package factom

// Heights contains all of the distinct heights for a factomd node and the
// Factom network.
type Heights struct {
	// The current directory block height of the local factomd node.
	DirectoryBlock uint64 `json:"directoryblockheight"`

	// The current block being worked on by the leaders in the network.
	// This block is not yet complete, but all transactions submitted will
	// go into this block (depending on network conditions, the transaction
	// may be delayed into the next block)
	Leader uint64 `json:"leaderheight"`

	// The height at which the factomd node has all the entry blocks.
	// Directory blocks are obtained first, entry blocks could be lagging
	// behind the directory block when syncing.
	EntryBlock uint64 `json:"entryblockheight"`

	// The height at which the local factomd node has all the entries. If
	// you added entries at a block height above this, they will not be
	// able to be retrieved by the local factomd until it syncs further.
	Entry uint64 `json:"entryheight"`
}

// Get uses c to call the "heights" RPC method and populates h with the result.
func (h *Heights) Get(c *Client) error {
	if err := c.FactomdRequest("heights", nil, &h); err != nil {
		return err
	}
	return nil
}
