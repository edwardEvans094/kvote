package main

import (
	"fmt"
	"log"
	"math/big"
	"os"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/joho/godotenv"

	blockchainVote "github.com/kvote/service"
	tb "gopkg.in/tucnak/telebot.v2"
)

var mCommand map[string]string
var mCreate map[string][]string

var mTx map[string]*tb.Message

// Bot object
type Bot struct {
	bot *tb.Bot
	// storage  *QuestionStorage
	deadline    int64
	voteNetwork *blockchainVote.Ethereum
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
		b.bot.Reply(m, `Em chào đại ca, đại ca muốn tạo Poll à, đưa em title nào`)
	} else if len(pollData) == 1 {
		b.bot.Reply(m, `Title dài vler, thế Options là gì thế ?`)
	} else {
		optionIndex := len(pollData)
		b.bot.Reply(m, fmt.Sprintf("Okey, option %d tiếp theo là gì ?", optionIndex))
	}

	fmt.Println(pollData)
	fmt.Println(mCreate[fmt.Sprintf("%d_%d", m.Chat.ID, m.Sender.ID)])

}

func (b Bot) handleSubmit(m *tb.Message) {
	pollData := mCreate[fmt.Sprintf("%d_%d", m.Chat.ID, m.Sender.ID)]
	fmt.Println(pollData)

	var title [32]byte
	optionList := make([][32]byte, len(pollData)-1)
	urlList := make([][32]byte, len(pollData)-1)
	var option, url [32]byte
	for i, d := range pollData {
		if i == 0 {
			//tiitle
			copy(title[:], d)
		} else {
			copy(option[:], d)
			copy(url[:], d)
			optionList = append(optionList, option)
			urlList = append(urlList, url)
		}
	}

	whiteListAddresses := []string{"0xd1263bec4e244d387f3205f6967cd68254c9a185", "0x2262d4f6312805851e3b27c40db2c7282e6e4a49"}
	voteData, encodeErr := b.voteNetwork.EncodeCreateCampaign(title, optionList, urlList, big.NewInt(999999999999), false, whiteListAddresses)
	if encodeErr != nil {
		panic(encodeErr)
	}

	// fmt.Println("===================vote data encoded: ", voteData, m.Text)
	txHash, errSendTx := b.voteNetwork.SendTx(m.Text, voteData)
	fmt.Println("===================txhash created: ", txHash, strings.ToLower(txHash))

	if errSendTx != nil {
		panic(errSendTx)
	}
	mTx[strings.ToLower(txHash)] = m
	b.bot.Send(m.Chat, fmt.Sprintf("Em đã tạo poll rồi nhé :v, tsHash đây: %s . Khi nào được em báo nhé!", txHash))
	// fmt.Println("********************* send trasaction done: ", txHash)
}

func (b Bot) handleDefault(m *tb.Message) {
	if m.Private() {
		b.bot.Send(m.Chat, `Nói clgv, éo hiểu!!`)
	}
}

func (b Bot) handleText(m *tb.Message) {
	switch mCommand[fmt.Sprintf("%d_%d", m.Chat.ID, m.Sender.ID)] {
	case "createPoll":
		b.handleCreatePoll(m)
	case "done":
		b.handleSubmit(m)
	default:
		b.handleDefault(m)
	}
}

func (b Bot) handleDone(m *tb.Message) {
	updateCurrentCommand("done", m)
	b.bot.Reply(m, "nhập passphrase: ")
	// inlineKeys := [][]tb.InlineButton{}

	// if len(pollData) > 0 {
	// 	for n := 1; n < len(pollData); n++ {
	// 		fmt.Println(pollData[n])
	// 		inlineBtn := tb.InlineButton{
	// 			Unique: fmt.Sprintf("%d", n),
	// 			Text:   pollData[n],
	// 		}
	// 		b.bot.Handle(&inlineBtn, func(c *tb.Callback) {
	// 			fmt.Println("--------------", c)
	// 			// option := 0
	// 			// for i, v := range questionOptions {
	// 			// 	if v == replyBtn.Text {
	// 			// 		option = i
	// 			// 	}
	// 			// }
	// 			// b.handleAnswer(c)
	// 			b.bot.Respond(c, &tb.CallbackResponse{
	// 				Text: fmt.Sprintf("fuck you, mày vừa chọn phương án %d phải không ?", n-1),
	// 			})

	// 		})
	// 		inlineKeysRow := []tb.InlineButton{inlineBtn}
	// 		inlineKeys = append(inlineKeys, inlineKeysRow)
	// 	}

	// 	// for index, value := range textArr {
	// 	// inlineBtn := tb.InlineButton{
	// 	// 	Unique: "",
	// 	// 	Text: value,
	// 	// }
	// 	// I
	// 	// 	inlineKeys[index][0].Text = textArr[index+1]
	// 	// }
	// }

	// b.bot.Send(m.Sender, "Hello!", &tb.ReplyMarkup{
	// 	// ReplyKeyboard: replyKeys,
	// 	InlineKeyboard: inlineKeys,
	// })
}

// func (b Bot) handleAnswer(c *tb.Callback) {
// 	// b.bot.Send(m.Sender, fmt.Sprintf("Mày vừa chọn phương án %d phải không", option))
// 	b.bot.Respond(c, &tb.CallbackResponse{...})
// }
func (b Bot) handleEventLog(vLog types.Log) {
	txHash := vLog.TxHash.Hex()
	fmt.Println(" on handle event log______________tx Hash____________", vLog.TxHash.Hex())
	// fmt.Println("__________________________", vLog.Data, fmt.Sprintf("%s", vLog.Data)) // pointer to event log
	campaignId := common.Bytes2Hex(vLog.Data)
	fmt.Println("__________________________", campaignId) // pointer to event log
	fmt.Println("__________________________address ", vLog.Address.Hex())
	fmt.Println("_______________toppic ", vLog.Topics)
	for _, topic := range vLog.Topics {
		fmt.Println("_______________toppic ", topic.Hex())
	}

	txChat := mTx[strings.ToLower(txHash)]
	if txChat != nil {
		b.bot.Send(txChat.Chat, fmt.Sprintf("Poll hash %s đã được mine trên network rồi nhé !", txHash))
	}
}

func (b Bot) subcribeEventLog(logs chan types.Log) {
	go func() {
		for {
			select {
			// case err := <-sub.Err():
			// 	log.Fatal(err)
			case vLog := <-logs:
				b.handleEventLog(vLog)
			}
		}
	}()
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	mTx = map[string]*tb.Message{}
	eventLogs := make(chan types.Log)

	voteNetwork, createNetworkErr := blockchainVote.NewEthereum(os.Getenv("VOTE_CONTRACT_RINKEBY"), os.Getenv("VOTE_ABI"), eventLogs)
	if createNetworkErr != nil {
		panic(createNetworkErr)
	}

	b, err := tb.NewBot(tb.Settings{
		Token: "454550958:AAHd8FQnm-x6uHjcIIKBHOfQhmh6TqRtsBY",
		Poller: &tb.LongPoller{
			Timeout: 10 * time.Second,
		},
	})

	mybot := Bot{
		bot: b,
		// storage:  storage,
		// deadline: botConfig.Deadline,
		voteNetwork: voteNetwork,
	}
	mybot.subcribeEventLog(eventLogs)

	if err != nil {
		log.Fatal(err)
		return
	}

	mybot.bot.Handle("/hello", func(m *tb.Message) {
		b.Send(m.Sender, "Hello world")
	})

	mybot.bot.Handle("/createPoll", func(m *tb.Message) {
		updateCurrentCommand("createPoll", m)
		// mybot.handleCreatePoll(m)
		b.Reply(m, "Em chào đại ca, đại ca muốn tạo Poll à, đưa em title nào")
	})

	mybot.bot.Handle("/done", func(m *tb.Message) {
		mybot.handleDone(m)
	})

	mybot.bot.Handle(tb.OnText, func(m *tb.Message) {
		// all the text messages that weren't
		// captured by existing handlers
		mybot.handleText(m)
		// fmt.Println("in handle text, %d, %d", m.Chat.ID, m.Sender.ID)
	})

	// b.Start()

	mybot.voteNetwork.GetCampaignData(big.NewInt(6))

}

// todo monitor tx status by hash
