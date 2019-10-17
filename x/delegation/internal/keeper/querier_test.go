package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"

	codec "github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/delegation/internal/keeper"
	"github.com/cosmos/cosmos-sdk/x/delegation/internal/types"
)

func TestQuery(t *testing.T) {
	input := setupTestInput()
	ctx := input.ctx
	k := input.dk

	cdc := codec.New()
	types.RegisterCodec(cdc)

	// some helpers
	grant1 := types.FeeAllowanceGrant{
		Granter: addr,
		Grantee: addr3,
		Allowance: &types.BasicFeeAllowance{
			SpendLimit: sdk.NewCoins(sdk.NewInt64Coin("atom", 555)),
			Expiration: types.ExpiresAtHeight(334455),
		},
	}
	grant2 := types.FeeAllowanceGrant{
		Granter: addr2,
		Grantee: addr3,
		Allowance: &types.BasicFeeAllowance{
			SpendLimit: sdk.NewCoins(sdk.NewInt64Coin("eth", 123)),
			Expiration: types.ExpiresAtHeight(334455),
		},
	}

	// let's set up some initial state here
	err := k.DelegateFeeAllowance(ctx, grant1)
	require.NoError(t, err)
	err = k.DelegateFeeAllowance(ctx, grant2)
	require.NoError(t, err)

	// now try some queries
	cases := map[string]struct {
		path  []string
		valid bool
		res   []types.FeeAllowanceGrant
	}{
		"bad path": {
			path: []string{"foo", "bar"},
		},
		"no data": {
			// addr in bech32
			path:  []string{"fees", "cosmos157ez5zlaq0scm9aycwphhqhmg3kws4qusmekll"},
			valid: true,
		},
		"two grants": {
			// addr3 in bech32
			path:  []string{"fees", "cosmos1qk93t4j0yyzgqgt6k5qf8deh8fq6smpn3ntu3x"},
			valid: true,
			res:   []types.FeeAllowanceGrant{grant1, grant2},
		},
	}

	querier := keeper.NewQuerier(k)
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			bz, err := querier(ctx, tc.path, abci.RequestQuery{})
			if !tc.valid {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			var grants []types.FeeAllowanceGrant
			serr := cdc.UnmarshalJSON(bz, &grants)
			require.NoError(t, serr)

			assert.Equal(t, tc.res, grants)
		})
	}

}