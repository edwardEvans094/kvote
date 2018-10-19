package blockchainVote

import (
	"log"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

// type RateWrapper struct {
// 	ExpectedRate []*big.Int `json:"expectedRate"`
// 	SlippageRate []*big.Int `json:"slippageRate"`
// }

type Ethereum struct {
	network    string
	networkAbi abi.ABI
	// tradeTopic       string
	// wrapper          string
	// wrapperAbi       abi.ABI
	// averageBlockTime int64
}

func NewEthereum(network string, abiString string) (*Ethereum, error) {
	networkAbi, err := abi.JSON(strings.NewReader(abiString))
	if err != nil {
		log.Print(err)
		return nil, err
	}

	ethereum := &Ethereum{
		network, networkAbi,
		// tradeTopic, wrapper,
		// wrapperAbi, averageBlockTime,
	}

	return ethereum, nil
}

func (self *Ethereum) EncodeCreateCampaign(title [32]byte, optionNames [][32]byte, optionUrls [][32]byte, end *big.Int, isMultipleChoices bool, whitelistedAddresses []string) (string, error) {

	// optionNameList := make([]string, 0)
	// for _, optionItem := range optionNames {
	// 	optionNameList = append(optionNameList, [32]byte(optionItem))
	// }

	// optionUrlList := make([]string, 0)
	// for _, optionUrl := range optionNames {
	// 	optionUrlList = append(optionUrlList, common.Hex2Bytes(optionUrl))
	// }

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
