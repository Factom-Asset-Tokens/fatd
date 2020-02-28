package state

import (
	"context"
	"fmt"
	"path/filepath"
	goruntime "runtime"

	"crawshaw.io/sqlite"
	"crawshaw.io/sqlite/sqlitex"
	jsonrpc2 "github.com/AdamSLevy/jsonrpc2/v14"
	"github.com/Factom-Asset-Tokens/factom"
	"github.com/Factom-Asset-Tokens/factom/fat104"
	"github.com/Factom-Asset-Tokens/fatd/internal/db"
	"github.com/Factom-Asset-Tokens/fatd/internal/db/contract"
	"github.com/Factom-Asset-Tokens/fatd/internal/runtime"
	"github.com/wasmerio/go-ext-wasm/wasmer"
	"golang.org/x/sync/errgroup"
)

type ContractResolver struct {
	ctx    context.Context
	cancel func()
	g      *errgroup.Group

	c *factom.Client

	conn *sqlite.Conn
	pool *sqlitex.Pool

	read     chan contractRequest
	download chan contractRequest
	write    chan contractWriteRequest
}

func NewContractResolver(ctx context.Context, c *factom.Client,
	dbURI string) (*ContractResolver, context.Context, error) {

	conn, pool, err := db.OpenConnPoolContract(ctx,
		filepath.Join(dbURI, "contract.sqlite3"))
	if err != nil {
		return nil, nil, err
	}

	ctx, cancel := context.WithCancel(ctx)
	g, ctx := errgroup.WithContext(ctx)
	n := goruntime.NumCPU()
	rsv := ContractResolver{ctx, cancel, g, c, conn, pool,
		make(chan contractRequest, n),
		make(chan contractRequest, n),
		make(chan contractWriteRequest, n),
	}

	// Concurrent reads and downloading.
	for i := 0; i < n; i++ {
		g.Go(rsv.reader)
		g.Go(rsv.downloader)
	}
	g.Go(rsv.writer) // Single writer.

	return &rsv, ctx, nil
}

func (rsv *ContractResolver) Close() {
	rsv.cancel()
	close(rsv.read)
	rsv.g.Wait()
	rsv.pool.Close()
	rsv.conn.Close()
}

func (rsv *ContractResolver) reader() error {
	conn := rsv.pool.Get(rsv.ctx)
	defer rsv.pool.Put(conn)
	for {
		select {
		case req, ok := <-rsv.read:
			if !ok {
				return nil
			}
			err := rsv.handleReadReq(conn, req)
			if err != nil {
				return err
			}
		case <-rsv.ctx.Done():
			return rsv.ctx.Err()
		}
	}
}
func (rsv *ContractResolver) handleReadReq(
	conn *sqlite.Conn, req contractRequest) error {

	var res contractResponse
	var id int64
	var err error
	switch req.Type {
	case contractRequestTypeValid:
		res.Valid, id, err = contract.SelectValid(conn, req.ChainID)
		if err != nil {
			return fmt.Errorf("contract.SelectValid(): %w", err)
		}
	case contractRequestTypeABIFunc:
		res.Func, id, err = contract.SelectABIFunc(conn, req.ChainID, req.FName)
		if err != nil {
			return fmt.Errorf("contract.SelectABIFunc(): %w", err)
		}
		res.Valid = res.Func != nil
		fallthrough
	case contractRequestTypeModule:
		res.Module, id, err = contract.SelectByChainID(conn, req.ChainID)
		if err != nil {
			return fmt.Errorf("contract.SelectByChainID(): %w", err)
		}
		res.Valid = res.Module != nil
	}

	if id < 0 {
		// Unknown contract, forward to downloader.
		rsv.download <- req
		return nil
	}
	// Send back response and close the channel.
	req.res <- res
	close(req.res)
	return nil
}

func (rsv *ContractResolver) downloader() error {
	for {
		select {
		case req, ok := <-rsv.download:
			if !ok {
				return nil
			}
			if err := rsv.handleDownloadReq(req); err != nil {
				return err
			}
		case <-rsv.ctx.Done():
			return rsv.ctx.Err()
		}
	}
}
func (rsv *ContractResolver) handleDownloadReq(req contractRequest) error {
	var con fat104.Contract
	var err error
	if con, err = fat104.Lookup(rsv.ctx, rsv.c, req.ChainID); err != nil {
		if _, ok := err.(jsonrpc2.Error); ok {
			req.res <- contractResponse{}
			return nil
		}
		return fmt.Errorf("fat104.Lookup(): %w", err)
	}

	// Download contract
	if err := con.Get(rsv.ctx, rsv.c); err != nil {
		if _, ok := err.(jsonrpc2.Error); ok {
			req.res <- contractResponse{}
			return nil
		}
		return fmt.Errorf("fat104.Contract.Get(): %w", err)
	}

	// Compile
	var mod wasmer.Module
	if mod, err = wasmer.CompileWithGasMetering(con.Wasm); err != nil {
		req.res <- contractResponse{}
		rsv.write <- contractWriteRequest{con, nil}
		return nil
	}

	// instantiate
	var vm *runtime.VM
	if vm, err = runtime.NewVM(&mod); err != nil {
		req.res <- contractResponse{}
		rsv.write <- contractWriteRequest{con, nil}
		return nil
	}

	// TODO: Improve chain agnostic ABI validation
	var rCtx runtime.Context
	rCtx.Context = rsv.ctx
	if err = vm.ValidateABI(&rCtx, con.ABI); err != nil {
		req.res <- contractResponse{}
		rsv.write <- contractWriteRequest{con, nil}
		return nil
	}

	// serialize
	var compiled []byte
	if compiled, err = mod.Serialize(); err != nil {
		return fmt.Errorf("wasmer.Module.Serialize(): %w", err)
	}

	req.res <- contractResponse{Module: &mod, Valid: true}
	rsv.write <- contractWriteRequest{con, compiled}
	return nil
}

type contractWriteRequest struct {
	fat104.Contract
	compiled []byte
}

func (rsv *ContractResolver) writer() error {
	for {
		select {
		case req, ok := <-rsv.write:
			if !ok {
				return nil
			}
			if err := rsv.handleWriteReq(req); err != nil {
				return err
			}
		case <-rsv.ctx.Done():
			return rsv.ctx.Err()
		}
	}
}

func (rsv *ContractResolver) handleWriteReq(req contractWriteRequest) error {
	// TODO: handle double insert.
	if _, err := contract.Insert(rsv.conn,
		req.Contract, req.compiled); err != nil {
		return fmt.Errorf("contract.Insert(): %w", err)
	}
	return nil
}

type contractRequestType int

const (
	contractRequestTypeModule  contractRequestType = iota
	contractRequestTypeValid   contractRequestType = iota
	contractRequestTypeABIFunc contractRequestType = iota
)

type contractRequest struct {
	ChainID *factom.Bytes32
	FName   string
	Type    contractRequestType
	res     chan contractResponse
}
type contractResponse struct {
	*wasmer.Module
	Valid bool
	*fat104.Func
}

func (rsv *ContractResolver) GetModule(ctx context.Context,
	chainID *factom.Bytes32) (*wasmer.Module, error) {

	res, err := rsv.request(ctx, contractRequest{
		ChainID: chainID,
		Type:    contractRequestTypeModule,
	})
	return res.Module, err
}
func (rsv *ContractResolver) GetValid(ctx context.Context,
	chainID *factom.Bytes32) (bool, error) {

	res, err := rsv.request(ctx, contractRequest{
		ChainID: chainID,
		Type:    contractRequestTypeValid,
	})
	return res.Valid, err
}
func (rsv *ContractResolver) GetABIFunc(ctx context.Context,
	chainID *factom.Bytes32, fname string) (*wasmer.Module, *fat104.Func, error) {

	res, err := rsv.request(ctx, contractRequest{
		ChainID: chainID,
		Type:    contractRequestTypeABIFunc,
		FName:   fname,
	})
	return res.Module, res.Func, err
}

func (rsv *ContractResolver) request(ctx context.Context,
	req contractRequest) (contractResponse, error) {
	req.res = make(chan contractResponse)
	rsv.read <- req
	var res contractResponse
	select {
	case res = <-req.res:
	case <-ctx.Done():
	}
	return res, ctx.Err()

}
