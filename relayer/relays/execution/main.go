package execution

import (
	"context"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	"github.com/snowfork/go-substrate-rpc-client/v4/types"
	"github.com/snowfork/snowbridge/relayer/chain/ethereum"
	"github.com/snowfork/snowbridge/relayer/chain/parachain"
	"github.com/snowfork/snowbridge/relayer/contracts"
	"github.com/snowfork/snowbridge/relayer/crypto/sr25519"
	"golang.org/x/sync/errgroup"
)

type Relay struct {
	config  *Config
	keypair *sr25519.Keypair
}

func NewRelay(
	config *Config,
	keypair *sr25519.Keypair,
) *Relay {
	return &Relay{
		config:  config,
		keypair: keypair,
	}
}

func (r *Relay) Start(ctx context.Context, eg *errgroup.Group) error {
	paraconn := parachain.NewConnection(r.config.Sink.Parachain.Endpoint, r.keypair.AsKeyringPair())
	ethconn := ethereum.NewConnection(r.config.Source.Ethereum.Endpoint, nil)

	err := paraconn.Connect(ctx)
	if err != nil {
		return err
	}

	err = ethconn.Connect(ctx)
	if err != nil {
		return err
	}

	writer := parachain.NewParachainWriter(
		paraconn,
		r.config.Sink.Parachain.MaxWatchedExtrinsics,
	)

	err = writer.Start(ctx, eg)
	if err != nil {
		return err
	}

	headerCache, err := ethereum.NewHeaderBlockCache(
		&ethereum.DefaultBlockLoader{Conn: ethconn},
	)
	if err != nil {
		return err
	}

	address := common.HexToAddress(r.config.Source.Contracts.OutboundChannel)
	contract, err := contracts.NewOutboundChannel(address, ethconn.Client())
	if err != nil {
		return err
	}

	opts := bind.WatchOpts{
		Context: ctx,
	}

	messages := make(chan *contracts.OutboundChannelMessage)

	key, err := types.EncodeToBytes(r.config.Source.LaneID)
	if err != nil {
		return fmt.Errorf("encode to bytes: %w", err)
	}

	sub, err := contract.WatchMessage(&opts, messages, [][]byte{key})
	if err != nil {
		return fmt.Errorf("create subscription: %w", err)
	}
	defer sub.Unsubscribe()

	for {
		select {
		case <-ctx.Done():
			return nil
		case err := <-sub.Err():
			return fmt.Errorf("message subscription: %w", err)
		case outboundMsg := <-messages:
			// wait until light client is updated with execution headers
			for {
				executionHeaderState, err := writer.GetLastExecutionHeaderState()
				if err != nil {
					return fmt.Errorf("fetch last execution header state: %w", err)
				}
				if outboundMsg.Raw.BlockNumber <= executionHeaderState.BlockNumber {
					break
				}
				select {
				case <-ctx.Done():
					return nil
				case <-time.After(6 * time.Second):
				}
			}

			inboundMsg, err := r.makeInboundMessage(ctx, headerCache, outboundMsg)
			if err != nil {
				return fmt.Errorf("make outgoing message: %w", err)
			}

			err = writer.WriteToParachainAndWatch(ctx, "InboundChannel.submit", inboundMsg)
			if err != nil {
				return fmt.Errorf("write to parachain: %w", err)
			}
		}
	}
}



func (r *Relay) makeInboundMessage(
	ctx context.Context,
	headerCache *ethereum.HeaderCache,
	event *contracts.OutboundChannelMessage,
) (*parachain.Message, error) {
	receiptTrie, err := headerCache.GetReceiptTrie(ctx, event.Raw.BlockHash)
	if err != nil {
		log.WithFields(logrus.Fields{
			"blockHash":   event.Raw.BlockHash.Hex(),
			"blockNumber": event.Raw.BlockNumber,
			"txHash":      event.Raw.TxHash.Hex(),
		}).WithError(err).Error("Failed to get receipt trie for event")
		return nil, err
	}

	msg, err := ethereum.MakeMessageFromEvent(&event.Raw, receiptTrie)
	if err != nil {
		log.WithFields(logrus.Fields{
			"address":     event.Raw.Address.Hex(),
			"blockHash":   event.Raw.BlockHash.Hex(),
			"blockNumber": event.Raw.BlockNumber,
			"txHash":      event.Raw.TxHash.Hex(),
		}).WithError(err).Error("Failed to generate message from ethereum event")
		return nil, err
	}

	return msg, nil
}
