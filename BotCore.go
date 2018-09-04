package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/stels-cs/vk-api-tools"
	"log"
	"strconv"
	"strings"
)

type MessageListenerCallback func(msg *VkApi.CallbackMessage) *BotResponse
type AfterCallback func(msg *VkApi.CallbackMessage, response *BotResponse)
type BeforeCallback func(msg *VkApi.CallbackMessage) bool

type MessageListener struct {
	filter   *MessageFilter
	callback MessageListenerCallback
}

type BotCore struct {
	token string

	groupId         int
	groupName       string
	groupScreenName string

	LongPollSettings VkApi.P

	api   *VkApi.Api
	Queue *VkApi.RequestQueue

	Logger *log.Logger

	LongPool *VkApi.BotLongPollServer

	userDB *UserPoll

	beforeCallback []BeforeCallback
	afterCallback  []AfterCallback
	listeners      []MessageListener

	bindPeerId       map[int]MessageListener
	bindFromId       map[int]MessageListener
	bindFromIdPeerId map[string]MessageListener

	onSendMessage func()

	oldTsValue int
}

func CreateBotCore(token string) *BotCore {
	bot := BotCore{
		token:            token,
		LongPollSettings: defaultLongPollSettings,
		listeners:        make([]MessageListener, 0),
		beforeCallback:   make([]BeforeCallback, 0),
		afterCallback:    make([]AfterCallback, 0),
		bindPeerId:       make(map[int]MessageListener),
		bindFromId:       make(map[int]MessageListener),
		bindFromIdPeerId: make(map[string]MessageListener),
	}

	return &bot
}

func (bot *BotCore) Start() error {
	bot.Queue = VkApi.GetRequestQueue(bot.api, 20)
	bot.LongPool = VkApi.GetBotLongPollServer(bot.api, bot.Logger, bot.groupId)
	bot.LongPool.SetListener(bot.callbackEventListener)
	if bot.oldTsValue != 0 {
		bot.LongPool.Ts = bot.oldTsValue
	}
	bot.userDB = GetPoll(bot.api, bot.Logger)
	q := make(chan error, 1)
	l := make(chan error, 1)

	go func() {
		bot.Queue.Start()
		q <- nil
	}()

	go func() {
		l <- bot.LongPool.Start()
	}()

	err := <-q

	if err != nil {
		return err
	}

	err = <-l
	return err
}

func (bot *BotCore) Stop() {
	bot.Queue.Stop()
	bot.LongPool.Stop()
}

func (bot *BotCore) GetName() string {
	return "BotCore"
}

func (bot *BotCore) bootstrap() error {
	api := VkApi.CreateApi(bot.token, "5.80", VkApi.GetHttpTransport(), 30)
	bot.api = api

	group, err := api.Call("groups.getById", VkApi.P{"fields": "screen_name"})
	if err != nil {
		return err
	}

	groupId := group.QIntDef("0.id", 0)
	groupName := group.QStringDef("0.name", "DELETED")
	groupScreen := group.QStringDef("0.screen_name", "club")
	if groupId == 0 {
		return errors.New("Cant fetch group id, from groups.getById, response " + string(*group.Response))
	}

	bot.groupId = groupId
	bot.groupName = groupName
	bot.groupScreenName = groupScreen

	bot.LongPollSettings["group_id"] = strconv.Itoa(bot.groupId)
	bot.LongPollSettings["enabled"] = "1"

	res, err := api.Call("groups.setLongPollSettings", bot.LongPollSettings)

	if err != nil {
		return err
	}

	if res.IntDef(0) != 1 {
		return errors.New("Cant call groups.setLongPollSettings (result is not 1), perhaps token without manage rights. response " + string(*res.Response))
	}

	return nil
}

func (bot *BotCore) GetGroupView() string {
	return fmt.Sprintf("%s https://vk.com/club%d", bot.groupName, bot.groupId)
}

func (bot *BotCore) callbackEventListener(event *VkApi.CallbackEvent) {
	if event.IsMessage() {
		if msg, err := event.GetMessage(); err == nil {

			if msg.Out == 1 {
				return
			}

			msg.GroupId = bot.groupId

			bot.fireMessage(msg)
		} else {
			bot.Logger.Println("Error on getting message " + err.Error() + "\nfrom: " + string(event.Object))
		}
	}
}

func (bot *BotCore) GetPeerFromKey(peerId, fromId int) string {
	return strconv.Itoa(peerId) + "-" + strconv.Itoa(fromId)
}

func (bot *BotCore) fireMessage(message *VkApi.CallbackMessage) {
	for _, callback := range bot.beforeCallback {
		if callback(message) == false {
			return
		}
	}

	filterList := make([]MessageListener, 0, len(bot.listeners))

	if rule, has := bot.bindFromId[message.FromId]; has {
		filterList = append(filterList, rule)
	}

	if rule, has := bot.bindPeerId[message.FromId]; has {
		filterList = append(filterList, rule)
	}

	key := bot.GetPeerFromKey(message.PeerId, message.FromId)
	if rule, has := bot.bindFromIdPeerId[key]; has {
		filterList = append(filterList, rule)
	}

	filterList = append(filterList, bot.listeners...)
	for _, listener := range filterList {
		filter := listener.filter
		if filter.Pass(message) {
			res := listener.callback(message)
			if res != nil {
				bot.fireResponse(res, message)
				return
			}
		}
	}
}

func (bot *BotCore) OnMessage() *MessageFilter {
	return GetDefaultFilter(bot)
}

func (bot *BotCore) fireResponse(response *BotResponse, message *VkApi.CallbackMessage) {

	for _, callback := range bot.afterCallback {
		callback(message, response)
	}

	if response == nil {
		return
	}

	if response.Message != nil {
		go func() {
			r := <-bot.SendMessage(response.Message, message.PeerId)
			if r.Err != nil {
				bot.Logger.Println(r.Err.Error())
				if strings.Index(r.Err.Error(), "Flood control") != -1 {
					//bot.floodCount++
				}
			}
		}()
	}
}

func (bot *BotCore) BeforeMessage(callback BeforeCallback) {
	bot.beforeCallback = append(bot.beforeCallback, callback)
}

func (bot *BotCore) AfterMessage(callback AfterCallback) {
	bot.afterCallback = append(bot.afterCallback, callback)
}

func (bot *BotCore) SendMessage(message *BotMessage, peerId int) chan VkApi.RequestResult {
	params := VkApi.P{
		"peer_id": strconv.Itoa(peerId),
		"message": message.Message,
	}

	if message.RandomId != 0 {
		params["random_id"] = strconv.Itoa(message.RandomId)
	}

	if message.Lat != 0 && message.Long != 0 {
		params["lat"] = strconv.FormatFloat(message.Lat, 'f', -1, 64)
		params["long"] = strconv.FormatFloat(message.Long, 'f', -1, 64)
	}

	if message.StickerId != 0 {
		params["sticker_id"] = strconv.Itoa(message.StickerId)
	}

	if message.Attachment != nil {
		params["attachment"] = strings.Join(message.Attachment, ",")
	}

	if message.ForwardMessages != nil {
		params["forward_messages"] = idsToString(message.ForwardMessages)
	}

	if message.Keyboard != nil {
		kbdString, err := json.Marshal(message.Keyboard)
		if err == nil {
			params["keyboard"] = string(kbdString)
		} else {
			bot.Logger.Println("Cant marshal keyboard")
			bot.Logger.Println(err)
		}
	}

	if bot.onSendMessage != nil {
		bot.onSendMessage()
	}

	return bot.Queue.Call(VkApi.CreateMethod("messages.send", params))
}

func (bot *BotCore) OnSendMessage(callback func()) {
	bot.onSendMessage = callback
}

func (bot *BotCore) GetTsValue() int {
	return bot.LongPool.Ts
}

func (bot *BotCore) SetOldTsValue(value int) {
	bot.oldTsValue = value
}
