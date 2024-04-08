package execution

import (
	"context"
	"fmt"
	geth "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/snowfork/go-substrate-rpc-client/v4/types"
	"github.com/snowfork/snowbridge/relayer/relays/beacon/header/syncer/api"
	"github.com/snowfork/snowbridge/relayer/relays/beacon/store"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	gethtypes "github.com/ethereum/go-ethereum/core/types"
	log "github.com/sirupsen/logrus"
	"github.com/snowfork/snowbridge/relayer/chain/ethereum"
	"github.com/snowfork/snowbridge/relayer/chain/parachain"
	"github.com/snowfork/snowbridge/relayer/crypto/sr25519"
	"github.com/snowfork/snowbridge/relayer/relays/beacon/header"
	"golang.org/x/sync/errgroup"
)

type LogOperatorRegistered struct {
	Operator   common.Address
	OperatorId [32]byte
}

type LogOperatorDeregistered struct {
	Operator   common.Address
	OperatorId [32]byte
}

type LogOperatorStakeUpdate struct {
	OperatorId   [32]byte
	QuorumNumber uint8
	Stake        big.Int
}
type Relay struct {
	config              *Config
	keypair             *sr25519.Keypair
	paraconn            *parachain.Connection
	ethconn             *ethereum.Connection
	beaconHeader        *header.Header
	stakeRegistry       common.Address
	registryCoordinator common.Address
	blockstore          Blockstore
	defaultStartHeight  uint64
	eventSigs           [][]common.Hash
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
	ethconn := ethereum.NewConnection(&r.config.Source.Ethereum, nil)

	err := paraconn.Connect(ctx)
	if err != nil {
		return err
	}
	r.paraconn = paraconn

	err = ethconn.Connect(ctx)
	if err != nil {
		return err
	}
	r.ethconn = ethconn
	writer := parachain.NewParachainWriter(
		paraconn,
		r.config.Sink.Parachain.MaxWatchedExtrinsics,
		r.config.Sink.Parachain.MaxBatchCallSize,
	)

	bs, err := NewBlockstore("beacon_execution.block")
	r.blockstore = *bs
	r.eventSigs = [][]common.Hash{{
		crypto.Keccak256Hash([]byte("")),
		crypto.Keccak256Hash([]byte("")),
		crypto.Keccak256Hash([]byte("")),
	}}

	if err != nil {
		return err
	}

	err = writer.Start(ctx, eg)
	if err != nil {
		return err
	}
	r.registryCoordinator = common.HexToAddress(r.config.Source.Contracts.Gateway)
	r.stakeRegistry = common.HexToAddress(r.config.Source.Contracts.Gateway)

	store := store.New(r.config.Source.Beacon.DataStore.Location, r.config.Source.Beacon.DataStore.MaxEntries)
	store.Connect()
	defer store.Close()

	beaconAPI := api.NewBeaconClient(r.config.Source.Beacon.Endpoint, r.config.Source.Beacon.Spec.SlotsInEpoch)
	beaconHeader := header.New(
		writer,
		beaconAPI,
		r.config.Source.Beacon.Spec,
		&store,
	)
	r.beaconHeader = &beaconHeader

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(6 * time.Second):

			blockNumber, err := ethconn.Client().BlockNumber(ctx)
			if err != nil {
				return fmt.Errorf("get last block number: %w", err)
			}

			events, err := r.findEvents(ctx, blockNumber)
			if err != nil {
				return fmt.Errorf("find events: %w", err)
			}

			for _, ev := range events {
				fmt.Printf("", ev)
				inboundMsg, err := r.makeInboundMessage(ev)
				if err != nil {
					return fmt.Errorf("make outgoing message: %w", err)
				}
				logger := log.WithFields(log.Fields{
					"address":     ev.Address.Hex(),
					"blockHash":   ev.BlockHash.Hex(),
					"blockNumber": ev.BlockNumber,
					"txHash":      ev.TxHash.Hex(),
					"txIndex":     ev.TxIndex,
					"logIndex":    ev.Index,
				})

				err = writer.WriteToParachainAndWatch(ctx, "EthereumInboundQueue.submit", inboundMsg)
				if err != nil {
					logger.Error("inbound message fail to sent")
					return fmt.Errorf("write to parachain: %w", err)
				}
				/*				paraNonce, _ = r.fetchLatestParachainNonce()
								if paraNonce != ev.Nonce {
									logger.Error("inbound message sent but fail to execute")
									return fmt.Errorf("inbound message fail to execute")
								}*/
				logger.Info("inbound message executed successfully")
			}
		}
	}
}

const BlocksPerQuery = 512

func (r *Relay) findEvents(
	ctx context.Context,
	latestFinalizedBlockNumber uint64,

) ([]gethtypes.Log, error) {

	var allEvents []gethtypes.Log
	blockNumber := latestFinalizedBlockNumber

	beginInt, err := r.blockstore.TryLoadLatestBlock()
	if err != nil {
		return nil, fmt.Errorf("filter events: %w", err)
	}
	begin := beginInt.Uint64()
	if begin == 0 {
		begin = r.defaultStartHeight
	}

	for {
		log.Info("loop")
		finish := false
		end := begin + BlocksPerQuery
		if end > blockNumber {
			end = blockNumber
			finish = true
		}
		opts := bind.FilterOpts{
			Start:   begin,
			End:     &end,
			Context: ctx,
		}
		events, err := r.findEventsWithFilter(&opts)
		if err != nil {
			return nil, fmt.Errorf("filter events: %w", err)
		}
		if len(events) > 0 {
			allEvents = append(allEvents, events...)
		}
		begin = end
		if finish {
			break
		}
	}
	r.blockstore.StoreBlock(new(big.Int).SetUint64(blockNumber))
	return allEvents, nil
}

func (r *Relay) findEventsWithFilter(opts *bind.FilterOpts) ([]gethtypes.Log, error) {
	query := geth.FilterQuery{
		FromBlock: big.NewInt(int64(opts.Start)),
		ToBlock:   big.NewInt(int64(*opts.End)),
		Addresses: []common.Address{
			r.registryCoordinator,
			r.stakeRegistry,
		},
		Topics: r.eventSigs,
	}
	return r.ethconn.Client().FilterLogs(opts.Context, query)
}

func (r *Relay) makeInboundMessage(
	event gethtypes.Log,
) (*parachain.Message, error) {

	return makeMessageFromEvent(event)
}

func makeMessageFromEvent(ev gethtypes.Log) (*parachain.Message, error) {

	var convertedTopics []types.H256
	for _, topic := range ev.Topics {
		convertedTopics = append(convertedTopics, types.H256(topic))
	}
	m := parachain.Message{
		EventLog: parachain.EventLog{
			Address:     types.H160(ev.Address),
			Topics:      convertedTopics,
			Data:        ev.Data,
			BlockNumber: ev.BlockNumber,
			LogIndex:    uint32(ev.Index),
		},
	}

	return &m, nil

}
