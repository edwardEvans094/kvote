package blockchainVote

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"os"
	"strings"
	"time"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

// type RateWrapper struct {
// 	ExpectedRate []*big.Int `json:"expectedRate"`
// 	SlippageRate []*big.Int `json:"slippageRate"`
// }

type Ethereum struct {
	network    string
	networkAbi abi.ABI
	rpc        *rpc.Client
	client     *ethclient.Client
}

func NewEthereum(network string, abiString string, eventLogsChannel chan types.Log) (*Ethereum, error) {
	networkAbi, err := abi.JSON(strings.NewReader(abiString))
	if err != nil {
		log.Print(err)
		return nil, err
	}

	rpc, rpcErr := rpc.Dial("wss://rinkeby.infura.io/ws")
	if rpcErr != nil {
		panic(rpcErr)
	}
	client := ethclient.NewClient(rpc)

	query := ethereum.FilterQuery{
		Addresses: []common.Address{common.HexToAddress(os.Getenv("VOTE_CONTRACT_RINKEBY"))},
	}

	// logs := make(chan types.Log)
	_, subErr := client.SubscribeFilterLogs(context.Background(), query, eventLogsChannel)
	if subErr != nil {
		panic(subErr)
	}

	ethereumIns := &Ethereum{
		network, networkAbi,
		rpc,
		client,
	}

	return ethereumIns, nil
}

func (self *Ethereum) EncodeCreateCampaign(title [32]byte, optionNames [][32]byte, optionUrls [][32]byte, end *big.Int, isMultipleChoices bool, whitelistedAddresses []string) (string, error) {
	listAddress := make([]common.Address, 0)
	for _, wAddress := range whitelistedAddresses {
		listAddress = append(listAddress, common.HexToAddress(wAddress))
	}
	encodedData, err := self.networkAbi.Pack("createCampaign", title, optionNames, optionUrls, end, isMultipleChoices, listAddress)
	if err != nil {
		// log.Print(err)
		return "", err
	}

	return common.Bytes2Hex(encodedData), nil
}

func (self *Ethereum) SendTx(passphrase string, voteData string) (string, error) {
	// ***************** unlock keystore
	d := time.Now().Add(5000 * time.Millisecond)
	ctx, cancel := context.WithDeadline(context.Background(), d)
	defer cancel()

	keyJson, readErr := ioutil.ReadFile("./bot.keystore")
	if readErr != nil {
		fmt.Println("key json read error:")
		// panic(readErr)
		return "", readErr
	}

	// Get the private key
	unlockedKey, keyErr := keystore.DecryptKey(keyJson, passphrase)
	if keyErr != nil {
		panic(keyErr)
	}
	// ***************** create data tx
	fmt.Println("---------------- addres: ", unlockedKey.Address.Hex())

	nonce, noneErr := self.client.NonceAt(ctx, unlockedKey.Address, nil)
	if noneErr != nil {
		// panic(noneErr)
		return "", noneErr
	}
	fmt.Println("===================nonce fetched: ", nonce)

	tx := types.NewTransaction(
		nonce,
		common.HexToAddress(os.Getenv("VOTE_CONTRACT_RINKEBY")),
		big.NewInt(0),
		500000,
		big.NewInt(50000000000),
		common.Hex2Bytes(voteData),
	)

	signTx, signErr := types.SignTx(tx, types.HomesteadSigner{}, unlockedKey.PrivateKey)
	if signErr != nil {
		// panic(signErr)
		return "", signErr
	}

	// *************** send tx
	errSendTransaction := self.client.SendTransaction(ctx, signTx)

	// ***************

	return signTx.Hash().String(), errSendTransaction
}
