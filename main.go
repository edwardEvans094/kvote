package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/joho/godotenv"

	blockchainVote "github.com/kvote/service"
	tb "gopkg.in/tucnak/telebot.v2"
)

var mCommand map[string]string
var mCreate map[string][]string

// Bot object
type Bot struct {
	bot *tb.Bot
	// storage  *QuestionStorage
	deadline int64
}

type Option struct {
	id     string
	name   string
	url    string
	voters []string
}

type Poll struct {
	id               string
	title            string
	options          []Option
	end              int
	admin            string
	creatorId        string
	isMultipleChoice bool
}

func updateCurrentCommand(command string, m *tb.Message) {
	if len(mCommand) == 0 {
		mCommand = map[string]string{}
	}
	mCommand[fmt.Sprintf("%d_%d", m.Chat.ID, m.Sender.ID)] = command
}

func (b Bot) handleCreatePoll(m *tb.Message) {
	if len(mCreate) == 0 {
		mCreate = map[string][]string{}
	}
	pollData := mCreate[fmt.Sprintf("%d_%d", m.Chat.ID, m.Sender.ID)]
	pollData = append(pollData, m.Text)

	mCreate[fmt.Sprintf("%d_%d", m.Chat.ID, m.Sender.ID)] = pollData

	if len(pollData) == 0 {
		b.bot.Reply(m, `đúng rồi đấy, tao đang tạo poll, đưa bố title nào`)
	} else if len(pollData) == 1 {
		b.bot.Reply(m, `DM title dài vler, thế Options đâu ?`)
	} else {
		optionIndex := len(pollData)
		b.bot.Reply(m, fmt.Sprintf("Okey, option %d của mày là gì ?", optionIndex))
	}

	fmt.Println(pollData)
	fmt.Println(mCreate[fmt.Sprintf("%d_%d", m.Chat.ID, m.Sender.ID)])

}

func (b Bot) handleDefault(m *tb.Message) {
	if m.Private() {
		b.bot.Send(m.Chat, `Mày nói clgv, éo hiểu!!`)
	}
}

func (b Bot) handleText(m *tb.Message) {
	switch mCommand[fmt.Sprintf("%d_%d", m.Chat.ID, m.Sender.ID)] {
	case "createPoll":
		b.handleCreatePoll(m)
	default:
		b.handleDefault(m)
	}
}

func (b Bot) handleDone(m *tb.Message) {

	inlineKeys := [][]tb.InlineButton{}

	pollData := mCreate[fmt.Sprintf("%d_%d", m.Chat.ID, m.Sender.ID)]
	fmt.Println(pollData)
	if len(pollData) > 0 {
		for n := 1; n < len(pollData); n++ {
			fmt.Println(pollData[n])
			inlineBtn := tb.InlineButton{
				Unique: fmt.Sprintf("%d", n),
				Text:   pollData[n],
			}
			b.bot.Handle(&inlineBtn, func(c *tb.Callback) {
				fmt.Println("--------------", c)
				// option := 0
				// for i, v := range questionOptions {
				// 	if v == replyBtn.Text {
				// 		option = i
				// 	}
				// }
				// b.handleAnswer(c)
				b.bot.Respond(c, &tb.CallbackResponse{
					Text: fmt.Sprintf("fuck you, mày vừa chọn phương án %d phải không ?", n-1),
				})

			})
			inlineKeysRow := []tb.InlineButton{inlineBtn}
			inlineKeys = append(inlineKeys, inlineKeysRow)
		}

		// for index, value := range textArr {
		// inlineBtn := tb.InlineButton{
		// 	Unique: "",
		// 	Text: value,
		// }
		// I
		// 	inlineKeys[index][0].Text = textArr[index+1]
		// }
	}

	b.bot.Send(m.Sender, "Hello!", &tb.ReplyMarkup{
		// ReplyKeyboard: replyKeys,
		InlineKeyboard: inlineKeys,
	})
}

// func (b Bot) handleAnswer(c *tb.Callback) {
// 	// b.bot.Send(m.Sender, fmt.Sprintf("Mày vừa chọn phương án %d phải không", option))
// 	b.bot.Respond(c, &tb.CallbackResponse{...})
// }

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// b, err := tb.NewBot(tb.Settings{
	// 	Token: "454550958:AAHd8FQnm-x6uHjcIIKBHOfQhmh6TqRtsBY",
	// 	Poller: &tb.LongPoller{
	// 		Timeout: 10 * time.Second,
	// 	},
	// })

	// mybot := Bot{
	// 	bot: b,
	// 	// storage:  storage,
	// 	// deadline: botConfig.Deadline,
	// }

	// if err != nil {
	// 	log.Fatal(err)
	// 	return
	// }

	// mybot.bot.Handle("/hello", func(m *tb.Message) {
	// 	b.Send(m.Sender, "Hello world")
	// })

	// mybot.bot.Handle("/createPoll", func(m *tb.Message) {
	// 	updateCurrentCommand("createPoll", m)
	// 	// mybot.handleCreatePoll(m)
	// 	b.Reply(m, "đúng rồi đấy, tao đang tạo poll, đưa bố title nào")
	// })

	// mybot.bot.Handle("/done", func(m *tb.Message) {
	// 	mybot.handleDone(m)
	// })

	// mybot.bot.Handle(tb.OnText, func(m *tb.Message) {
	// 	// all the text messages that weren't
	// 	// captured by existing handlers
	// 	mybot.handleText(m)
	// 	// fmt.Println("in handle text, %d, %d", m.Chat.ID, m.Sender.ID)
	// })

	// b.Start()

	fmt.Println("===================START: ", os.Getenv("VOTE_CONTRACT_RINKEBY"))
	// ***************** endcode data
	var title, option, url [32]byte
	copy(option[:], "option title")
	copy(url[:], "https://stackoverflow.com")
	copy(title[:], "vote title")

	listOptionsName := [][32]byte{option, option}
	listOptionsUrl := [][32]byte{url, url}
	voteNetwork, createNetworkErr := blockchainVote.NewEthereum(os.Getenv("VOTE_CONTRACT_RINKEBY"), os.Getenv("VOTE_ABI"))
	if createNetworkErr != nil {
		panic(createNetworkErr)
	}
	voteData, encodeErr := voteNetwork.EncodeCreateCampaign(title, listOptionsName, listOptionsUrl, big.NewInt(1539839022), false, nil)
	if encodeErr != nil {
		panic(encodeErr)
	}

	fmt.Println("===================vote data encoded: ", voteData)

	// ***************** unlock keystore
	d := time.Now().Add(5000 * time.Millisecond)
	ctx, cancel := context.WithDeadline(context.Background(), d)
	defer cancel()

	rpc, rpcErr := rpc.DialHTTP("https://rinkeby.infura.io")
	if rpcErr != nil {
		panic(rpcErr)
	}
	client := ethclient.NewClient(rpc)

	keyJson, readErr := ioutil.ReadFile("./bot.keystore")
	if readErr != nil {
		fmt.Println("key json read error:")
		panic(readErr)
	}

	// Get the private key
	unlockedKey, keyErr := keystore.DecryptKey(keyJson, "123qwe123qwe")
	if keyErr != nil {
		panic(keyErr)
	}
	fmt.Println("===================keystore unlocked: ", unlockedKey)
	// ***************** create data tx
	fmt.Println("---------------- addres: ", unlockedKey.Address.Hex())

	nonce, noneErr := client.NonceAt(ctx, unlockedKey.Address, nil)
	if noneErr != nil {
		panic(noneErr)
	}
	fmt.Println("===================nonce fetched: ", nonce)

	tx := types.NewTransaction(
		nonce,
		common.HexToAddress(os.Getenv("VOTE_CONTRACT_RINKEBY")),
		big.NewInt(0),
		500000,
		big.NewInt(50000000000),
		[]byte(voteData),
	)
	fmt.Println("===================tx created: ", tx, unlockedKey)
	// ************** sign data
	// signTx, signErr := types.SignTx(tx, types.NewEIP155Signer(big.NewInt(1)), unlockedKey.PrivateKey)
	signTx, signErr := types.SignTx(tx, types.HomesteadSigner{}, unlockedKey.PrivateKey)
	if signErr != nil {
		panic(signErr)
	}

	fmt.Println("===================tx signed: ", signTx)

	// *************** send tx

	errSendTransaction := client.SendTransaction(ctx, signTx)
	if errSendTransaction != nil {
		panic(errSendTransaction)
	}

	fmt.Println("===================DONE ")
	fmt.Printf("tx sent: %s", tx.Hash().Hex())
	// ks := keystore.NewKeyStore(
	// 	"./bot.keystore",
	// 	keystore.LightScryptN,
	// 	keystore.LightScryptP)

	// botAccount := &Account{
	// 	Address: "0x6F0311366C7178A8bE4392347c82415D7298278e",
	// }
	// unlockedKey, _ := ks.Unlock(botAccount, "123qwe")
	// nonce, _ := client.NonceAt(ctx, unlockedKey.Address)

}
