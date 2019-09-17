package ante_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	errs "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto"
)

func TestSetup(t *testing.T) {
	// setup
	_, ctx := createTestApp(true)

	// keys and addresses
	priv1, _, addr1 := types.KeyTestPubAddr()

	// msg and signatures
	msg1 := types.NewTestMsg(addr1)
	fee := types.NewTestStdFee()

	msgs := []sdk.Msg{msg1}

	privs, accNums, seqs := []crypto.PrivKey{priv1}, []uint64{0}, []uint64{0}
	tx := types.NewTestTx(ctx, msgs, privs, accNums, seqs, fee)

	sud := ante.NewSetupDecorator()
	antehandler := sdk.ChainDecorators(sud)

	// Set height to non-zero value for GasMeter to be set
	ctx = ctx.WithBlockHeight(1)

	// Context GasMeter Limit not set
	require.Equal(t, uint64(0), ctx.GasMeter().Limit(), "GasMeter set with limit before setup")

	newCtx, err := antehandler(ctx, tx, false)
	require.Nil(t, err, "SetupDecorator returned error")

	// Context GasMeter Limit should be set after SetupDecorator runs
	require.Equal(t, fee.Gas, newCtx.GasMeter().Limit(), "GasMeter not set correctly")
}

func TestRecoverPanic(t *testing.T) {
	// setup
	_, ctx := createTestApp(true)

	// keys and addresses
	priv1, _, addr1 := types.KeyTestPubAddr()

	// msg and signatures
	msg1 := types.NewTestMsg(addr1)
	fee := types.NewTestStdFee()

	msgs := []sdk.Msg{msg1}

	privs, accNums, seqs := []crypto.PrivKey{priv1}, []uint64{0}, []uint64{0}
	tx := types.NewTestTx(ctx, msgs, privs, accNums, seqs, fee)

	sud := ante.NewSetupDecorator()
	antehandler := sdk.ChainDecorators(sud, OutOfGasDecorator{})

	// Set height to non-zero value for GasMeter to be set
	ctx = ctx.WithBlockHeight(1)

	newCtx, err := antehandler(ctx, tx, false)

	require.NotNil(t, err, "Did not return error on OutOfGas panic")

	require.True(t, errs.ErrOutOfGas.Is(err), "Returned error is not an out of gas error")
	require.Equal(t, fee.Gas, newCtx.GasMeter().Limit())

	antehandler = sdk.ChainDecorators(sud, PanicDecorator{})
	require.Panics(t, func() { antehandler(ctx, tx, false) }, "Recovered from non-Out-of-Gas panic")
}

type OutOfGasDecorator struct{}

// AnteDecorator that will throw OutOfGas panic
func (ogd OutOfGasDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	overLimit := ctx.GasMeter().Limit() + 1

	// Should panic with outofgas error
	ctx.GasMeter().ConsumeGas(overLimit, "test panic")

	// not reached
	return next(ctx, tx, simulate)
}

type PanicDecorator struct{}

func (pd PanicDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	panic("random error")

	// not reached
	return ctx, nil
}