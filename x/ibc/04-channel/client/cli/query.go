package cli

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	cli "github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/state"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/x/ibc/02-client"
	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/client/utils"
	"github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/merkle"
	"github.com/cosmos/cosmos-sdk/x/ibc/version"
)

const (
	FlagProve = "prove"
)

func object(ctx context.CLIContext, cdc *codec.Codec, storeKey string, version int64, connid, chanid string) channel.CLIObject {
	prefix := []byte("v" + strconv.FormatInt(version, 10))
	path := merkle.NewPath([][]byte{[]byte(storeKey)}, prefix)
	base := state.NewBase(cdc, sdk.NewKVStoreKey(storeKey), prefix)
	climan := client.NewManager(base)
	connman := connection.NewManager(base, climan)
	man := channel.NewManager(base, connman)
	return man.CLIQuery(ctx, path, connid, chanid)
}

func GetQueryCmd(storeKey string, cdc *codec.Codec) *cobra.Command {
	ibcQueryCmd := &cobra.Command{
		Use:                        "connection",
		Short:                      "Channel query subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
	}

	ibcQueryCmd.AddCommand(cli.GetCommands(
		GetCmdQueryChannel(storeKey, cdc),
	)...)
	return ibcQueryCmd
}

func QueryChannel(ctx context.CLIContext, obj channel.CLIObject, prove bool) (res utils.JSONObject, err error) {
	conn, connp, err := obj.Channel(ctx)
	if err != nil {
		return
	}
	avail, availp, err := obj.Available(ctx)
	if err != nil {
		return
	}
	/*
		kind, kindp, err := obj.Kind(ctx)
		if err != nil {
			return
		}
	*/
	seqsend, seqsendp, err := obj.SeqSend(ctx)
	if err != nil {
		return
	}

	seqrecv, seqrecvp, err := obj.SeqRecv(ctx)
	if err != nil {
		return
	}

	if prove {
		return utils.NewJSONObject(
			conn, connp,
			avail, availp,
			//			kind, kindp,
			seqsend, seqsendp,
			seqrecv, seqrecvp,
		), nil
	}

	return utils.NewJSONObject(
		conn, nil,
		avail, nil,
		seqsend, nil,
		seqrecv, nil,
	), nil
}

func GetCmdQueryChannel(storeKey string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "channel [chanid]",
		Short: "Query stored channel",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.NewCLIContext().WithCodec(cdc)
			obj := object(ctx, cdc, storeKey, version.Version, args[0], args[1])
			jsonobj, err := QueryChannel(ctx, obj, viper.GetBool(FlagProve))
			if err != nil {
				return err
			}

			fmt.Printf("%s\n", codec.MustMarshalJSONIndent(cdc, jsonobj))

			return nil
		},
	}

	cmd.Flags().Bool(FlagProve, false, "(optional) show proofs for the query results")

	return cmd
}
