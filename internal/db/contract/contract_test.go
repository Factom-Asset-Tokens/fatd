package contract_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"testing"

	"crawshaw.io/sqlite"
	"crawshaw.io/sqlite/sqlitex"
	"github.com/Factom-Asset-Tokens/factom"
	"github.com/Factom-Asset-Tokens/factom/fat104"
	"github.com/Factom-Asset-Tokens/factom/fat107"
	"github.com/Factom-Asset-Tokens/fatd/internal/db/address"
	"github.com/Factom-Asset-Tokens/fatd/internal/db/contract"
	"github.com/stretchr/testify/require"
	"github.com/wasmerio/go-ext-wasm/wasmer"
)

func TestContracts(t *testing.T) {
	require := require.New(t)
	conn, err := sqlite.OpenConn("", 0)
	require.NoError(err)
	defer conn.Close()

	require.NoError(sqlitex.ExecScript(conn,
		`ATTACH DATABASE 'file::memory:' AS "contract";`),
		"ATTACH DATABASE")

	require.NoError(sqlitex.ExecScript(conn, contract.CreateTable), "CreateTable")

	chainIDAdd, id, err := insertWasmFile(t, conn, "./testdata/add.wasm")
	require.NoError(err)
	require.Equal(int64(1), id, "Insert(): ./testdata/add.wasm")

	_, id, err = insertWasmFile(t, conn, "./testdata/add.wasm")
	require.Error(err, "sqlite.Stmt.Step: SQLITE_CONSTRAINT_UNIQUE")
	require.Equal(int64(-1), id, "Insert(): ./testdata/add.wasm: duplicate")

	_, id, err = insertWasmFile(t, conn, "./testdata/nop.wasm")
	require.NoError(err)
	require.Equal(int64(2), id, "Insert(): ./testdata/nop.wasm")

	chainIDInvalid, id, err := insertWasmFile(t, conn, "./testdata/invalid.txt")
	require.NoError(err)
	require.Equal(int64(3), id, "Insert(): ./testdata/invalid.txt")

	count, err := contract.SelectCount(conn, false)
	require.NoError(err)
	require.Equal(int64(3), count)

	count, err = contract.SelectCount(conn, true)
	require.NoError(err)
	require.Equal(int64(2), count)

	require.NoError(contract.Validate(conn))

	{
		mod, id, err := contract.SelectByChainID(conn, &chainIDAdd)
		require.NoError(err)
		require.NotNil(mod)
		require.Equal(id, int64(1))
		_, err = mod.Instantiate()
		require.NoError(err)
	}

	cached, err := contract.SelectIsCached(conn, 1)
	require.NoError(err)
	require.True(cached, "IsCached")

	require.NoError(contract.ClearCompiledCache(conn))

	cached, err = contract.SelectIsCached(conn, 1)
	require.NoError(err)
	require.False(cached, "IsCached")

	cached, err = contract.SelectIsCached(conn, -1)
	require.Error(err)
	require.False(cached, "IsCached")

	{
		mod, err := contract.SelectByID(conn, 2)
		require.NoError(err)
		require.NotNil(mod)
		_, err = mod.Instantiate()
		require.NoError(err)

		require.NoError(contract.Cache(conn, 2, mod))
		cached, err := contract.SelectIsCached(conn, 2)
		require.NoError(err)
		require.True(cached, "IsCached")
	}

	{
		mod, id, err := contract.SelectByChainID(conn, &chainIDInvalid)
		require.NoError(err)
		require.Nil(mod)
		require.Equal(int64(3), id)
	}

	release := sqlitex.Save(conn)

	blob, err := conn.OpenBlob("contract", "contract", "wasm", 2, true)
	require.NoError(err)

	_, err = blob.WriteAt([]byte("hello world"), 0)
	require.NoError(err)
	require.NoError(blob.Close())

	require.Error(contract.Validate(conn))

	err = fmt.Errorf("rollback")
	release(&err)

	release = sqlitex.Save(conn)

	blob, err = conn.OpenBlob("contract", "contract", "first_entry", 2, true)
	require.NoError(err)

	_, err = blob.WriteAt([]byte("hello"), 38)
	require.NoError(err)
	require.NoError(blob.Close())

	require.Error(contract.Validate(conn))

	err = fmt.Errorf("rollback")
	release(&err)

	require.NoError(sqlitex.ExecScript(conn, address.CreateTable),
		"address.CreateTable")

	fs, err := factom.GenerateFsAddress()
	require.NoError(err)
	fa := fs.FAAddress()
	adrID, err := address.Add(conn, &fa, 1000)
	require.NoError(err)

	require.NoError(sqlitex.ExecScript(conn, address.CreateTableContract),
		"address.CreateTableContract")

	require.NoError(address.InsertContract(conn, adrID, 1, &chainIDAdd))
	id, chainID, err := address.SelectContract(conn, adrID)
	require.NoError(err)
	require.EqualValues(1, id)
	require.Equal(chainIDAdd, chainID)
	require.NoError(address.DeleteContract(conn, adrID))
	id, chainID, err = address.SelectContract(conn, adrID)
	require.NoError(err)
	require.EqualValues(-1, id)
	require.True(chainID.IsZero())
}

func insertWasmFile(t *testing.T, conn *sqlite.Conn, fileName string) (
	factom.Bytes32, int64, error) {
	require := require.New(t)
	wasm, err := ioutil.ReadFile(fileName)
	require.NoError(err)

	var compiled []byte
	mod, err := wasmer.CompileWithGasMetering(wasm)
	if err == nil {
		defer mod.Close()

		compiled, err = mod.Serialize()
		require.NoError(err)
	}

	hash := factom.Bytes32(sha256.Sum256(wasm))
	hash = sha256.Sum256(hash[:])
	wasmBuf := bytes.NewBuffer(wasm)
	chainID, _, _, _, reveals, _, err := fat107.Generate(
		nil, factom.EsAddress{}, wasmBuf, nil, uint64(len(wasm)), &hash, nil)
	require.NoError(err)
	var first factom.Entry
	require.NoError(first.UnmarshalBinary(reveals[0]))
	reveals = nil // No need to keep in memory.

	var con fat104.Contract
	con.Entry = first
	con.Wasm = wasm

	require.NoError(json.Unmarshal(first.Content, &con))

	id, err := contract.Insert(conn, con, compiled)
	return chainID, id, err
}
