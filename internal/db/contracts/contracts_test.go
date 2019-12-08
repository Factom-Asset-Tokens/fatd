package contracts

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io/ioutil"
	"testing"

	"crawshaw.io/sqlite"
	"crawshaw.io/sqlite/sqlitex"
	"github.com/Factom-Asset-Tokens/factom"
	"github.com/Factom-Asset-Tokens/factom/fat107"
	"github.com/stretchr/testify/require"
	"github.com/wasmerio/go-ext-wasm/wasmer"
)

func TestContracts(t *testing.T) {
	require := require.New(t)
	conn, err := sqlite.OpenConn(":memory:", 0)
	require.NoError(err)
	defer conn.Close()

	require.NoError(sqlitex.ExecScript(conn, CreateTable), "CreateTable")

	chainIDAdd, id, err := insertWasmFile(t, conn, "./testdata/add.wasm")
	require.NoError(err)
	require.Equal(int64(1), id, "Insert(): ./testdata/add.wasm")

	_, id, err = insertWasmFile(t, conn, "./testdata/add.wasm")
	require.Error(err, "sqlite.Stmt.Step: SQLITE_CONSTRAINT_UNIQUE")
	require.Equal(int64(-1), id, "Insert(): ./testdata/add.wasm: duplicate")

	_, id, err = insertWasmFile(t, conn, "./testdata/nop.wasm")
	require.NoError(err)
	require.Equal(int64(2), id, "Insert(): ./testdata/nop.wasm")

	chainIDInvalid, id, err := insertWasmFile(t, conn, "./testdata/invalid.wasm")
	require.NoError(err)
	require.Equal(int64(3), id, "Insert(): ./testdata/invalid.wasm")

	count, err := SelectCount(conn, false)
	require.NoError(err)
	require.Equal(int64(3), count)

	count, err = SelectCount(conn, true)
	require.NoError(err)
	require.Equal(int64(2), count)

	require.NoError(Validate(conn))

	{
		mod, id, err := SelectByChainID(conn, &chainIDAdd)
		require.NoError(err)
		require.NotNil(mod)
		require.Equal(id, int64(1))
		_, err = mod.Instantiate()
		require.NoError(err)
	}

	cached, err := SelectIsCached(conn, 1)
	require.NoError(err)
	require.True(cached, "IsCached")

	require.NoError(ClearCompiledCache(conn))

	cached, err = SelectIsCached(conn, 1)
	require.NoError(err)
	require.False(cached, "IsCached")

	cached, err = SelectIsCached(conn, -1)
	require.Error(err)
	require.False(cached, "IsCached")

	{
		mod, err := SelectByID(conn, 2)
		require.NoError(err)
		require.NotNil(mod)
		_, err = mod.Instantiate()
		require.NoError(err)

		require.NoError(Cache(conn, 2, mod))
		cached, err := SelectIsCached(conn, 2)
		require.NoError(err)
		require.True(cached, "IsCached")
	}

	{
		mod, id, err := SelectByChainID(conn, &chainIDInvalid)
		require.NoError(err)
		require.Nil(mod)
		require.Equal(int64(3), id)
	}

	release := sqlitex.Save(conn)

	blob, err := conn.OpenBlob("", "contracts", "wasm", 2, true)
	require.NoError(err)

	_, err = blob.WriteAt([]byte("hello world"), 0)
	require.NoError(err)
	require.NoError(blob.Close())

	require.Error(Validate(conn))

	err = fmt.Errorf("rollback")
	release(&err)

	release = sqlitex.Save(conn)

	blob, err = conn.OpenBlob("", "contracts", "first_entry", 2, true)
	require.NoError(err)

	_, err = blob.WriteAt([]byte("hello"), 38)
	require.NoError(err)
	require.NoError(blob.Close())

	require.Error(Validate(conn))

	err = fmt.Errorf("rollback")
	release(&err)
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

	id, err := Insert(conn, first, wasm, compiled)
	return chainID, id, err
}
