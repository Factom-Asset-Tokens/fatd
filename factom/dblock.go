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

package factom

// DBlock represents a Factom Directory Block.
type DBlock struct {
	Height uint64 `json:"height"`

	// DBlock.Get populates EBlocks with their ChainID and KeyMR.
	EBlocks []EBlock `json:"dbentries,omitempty"`
}

// IsPopulated returns true if db has already been successfully populated by a
// call to Get. IsPopulated returns false if db.EBlocks is nil.
func (db DBlock) IsPopulated() bool {
	return db.EBlocks != nil
}

// Get queries factomd for the Directory Block at db.Height. After a successful
// call, the EBlocks will all have their ChainID and KeyMR, but not their
// Entries. Call Get on the EBlocks individually to populate their Entries.
func (db *DBlock) Get(c *Client) error {
	if db.IsPopulated() {
		return nil
	}

	result := struct{ DBlock *DBlock }{DBlock: db}
	if err := c.FactomdRequest("dblock-by-height", db, &result); err != nil {
		return err
	}

	return nil
}
