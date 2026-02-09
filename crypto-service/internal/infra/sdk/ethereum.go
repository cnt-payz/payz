package sdk

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math/big"
	"time"

	paymentpb "github.com/cnt-payz/payz/crypto-service/api/payment/v1"
	"github.com/cnt-payz/payz/crypto-service/config"
	walletusecase "github.com/cnt-payz/payz/crypto-service/internal/application/usecase"
	domainrepo "github.com/cnt-payz/payz/crypto-service/internal/domain/repository"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

var (
	ErrInvalidAddress = errors.New("invalid address")
)

const (
	usdtContractAddress = "0xc2132D05D31c914a87C6611C10748AEb04B58e8F"
	minConfirmations    = 10
)

type ethSDK struct {
	client               *ethclient.Client
	wsClient             *ethclient.Client
	dbRepo               domainrepo.DatabaseRepo
	paymentServiceClient paymentpb.PaymentClient
	log                  *slog.Logger
}

func NewETHClient(cfg *config.Config, paymentServiceClient paymentpb.PaymentClient, dbRepo domainrepo.DatabaseRepo, log *slog.Logger) (walletusecase.ETHSDK, error) {
	client, err := ethclient.Dial(cfg.Node.PolygonTestnetAddress)
	if err != nil {
		return nil, err
	}

	wsClient, err := ethclient.Dial(cfg.Node.PolygonTestnetWebsocketAddress)
	if err != nil {
		return nil, err
	}

	return &ethSDK{
		client:               client,
		wsClient:             wsClient,
		paymentServiceClient: paymentServiceClient,
		dbRepo:               dbRepo,
		log:                  log,
	}, nil
}

func (es *ethSDK) ConfirmDeposit(ctx context.Context, shopAddress, takerAddress string, amount uint64) error {
	shopAddr := common.HexToAddress(shopAddress)
	takerAddr := common.HexToAddress(takerAddress)

	if shopAddr == (common.Address{}) {
		return errors.New("invalid shop address")
	}
	if takerAddr == (common.Address{}) {
		return errors.New("invalid taker address")
	}

	usdtAddr := common.HexToAddress(usdtContractAddress)
	transferEventHash := crypto.Keccak256Hash([]byte("Transfer(address,address,uint256)"))

	toTopic := addressToTopic(shopAddr)
	expectedFrom := takerAddr

	es.log.Debug("subscribing to transfers",
		slog.String("shop addr", shopAddress),
		slog.String("taker addr", takerAddress),
		slog.Uint64("amount", amount),
	)

	query := ethereum.FilterQuery{
		Addresses: []common.Address{usdtAddr},
		Topics: [][]common.Hash{
			{transferEventHash},
			nil,
			{toTopic},
		},
	}

	logs := make(chan types.Log)
	sub, err := es.wsClient.SubscribeFilterLogs(ctx, query, logs)
	if err != nil {
		return fmt.Errorf("failed to subscribe: %v", err)
	}
	defer sub.Unsubscribe()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case err := <-sub.Err():
			return fmt.Errorf("subscription error: %v", err)
		case vLog := <-logs:
			from := common.HexToAddress(vLog.Topics[1].Hex())
			if from != expectedFrom {
				continue
			}

			receivedAmount := new(big.Int).SetBytes(vLog.Data).Uint64()
			if receivedAmount < amount {
				continue
			}

			if !es.waitForConfirmations(ctx, vLog.BlockNumber) {
				continue
			}

			es.log.Debug("deposit confirmed", slog.String("tx_hash", vLog.TxHash.Hex()))
			return nil
		}
	}
}

func (es *ethSDK) waitForConfirmations(ctx context.Context, blockNumber uint64) bool {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return false

		case <-ticker.C:
			header, err := es.client.HeaderByNumber(context.Background(), nil)
			if err != nil {
				es.log.Warn("failed to get header during confirmation wait")
				continue
			}

			currentBlock := header.Number.Uint64()
			confirmations := currentBlock - blockNumber
			if confirmations >= minConfirmations {
				return true
			}

			es.log.Debug("waiting for confirmations",
				slog.Uint64("tx block", blockNumber),
				slog.Uint64("current block", currentBlock),
				slog.Uint64("confirmations", confirmations),
				slog.Uint64("required", minConfirmations),
			)
		}
	}
}

func addressToTopic(addr common.Address) common.Hash {
	var hash common.Hash
	copy(hash[12:], addr[:])
	return hash
}

func (es *ethSDK) GetUSDTAmount(ctx context.Context, address string) (uint64, error) {
	walletAddr := common.HexToAddress(address)
	if walletAddr == (common.Address{}) {
		return 0, ErrInvalidAddress
	}

	usdtAddr := common.HexToAddress(usdtContractAddress)

	balanceOfSig := crypto.Keccak256Hash([]byte("balanceOf(address)"))

	data := make([]byte, 4+32)
	copy(data[0:4], balanceOfSig[:4])
	copy(data[4:36], common.LeftPadBytes(walletAddr.Bytes(), 32))

	msg := ethereum.CallMsg{
		To:   &usdtAddr,
		Data: data,
	}

	result, err := es.client.CallContract(ctx, msg, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to call balanceOf: %w", err)
	}

	balance := new(big.Int).SetBytes(result)

	return balance.Uint64(), nil
}

func (es *ethSDK) Withdraw(ctx context.Context, fromAddress, toAddress string, amount uint64) error {
	fromAddr := common.HexToAddress(fromAddress)
	toAddr := common.HexToAddress(toAddress)

	if fromAddr == (common.Address{}) {
		return errors.New("invalid from address")
	}
	if toAddr == (common.Address{}) {
		return errors.New("invalid to address")
	}

	usdtAddr := common.HexToAddress(usdtContractAddress)

	transferData := []byte{
		0xa9, 0x05, 0x9c, 0xbb,
	}
	transferData = append(transferData, common.LeftPadBytes(toAddr.Bytes(), 32)...)
	amountBytes := make([]byte, 32)
	big.NewInt(int64(amount)).FillBytes(amountBytes)
	transferData = append(transferData, amountBytes...)

	nonce, err := es.client.PendingNonceAt(ctx, fromAddr)
	if err != nil {
		return fmt.Errorf("failed to get nonce: %v", err)
	}

	gasPrice, err := es.client.SuggestGasPrice(ctx)
	if err != nil {
		return fmt.Errorf("failed to get gas price: %v", err)
	}

	gasLimit := uint64(65000)

	tx := types.NewTransaction(nonce, usdtAddr, big.NewInt(0), gasLimit, gasPrice, transferData)

	chainID, err := es.client.ChainID(ctx)
	if err != nil {
		return fmt.Errorf("failed to get chain ID: %v", err)
	}

	signedTx, err := es.dbRepo.SignTransaction(ctx, fromAddress, tx, chainID)
	if err != nil {
		return fmt.Errorf("failed to sign transaction: %v", err)
	}

	if err := es.client.SendTransaction(ctx, signedTx); err != nil {
		return fmt.Errorf("failed to send transaction: %v", err)
	}

	es.log.Debug("Withdrawal transaction sent",
		"from", fromAddress,
		"to", toAddress,
		"amount", amount,
		"tx hash", signedTx.Hash().Hex(),
	)

	return nil
}
