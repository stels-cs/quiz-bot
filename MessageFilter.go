package main

import (
	"github.com/stels-cs/vk-api-tools"
	"strings"
)

const ITrue = 1
const IFalse = -1
const IAny = 0

type MessageFilter struct {
	mention   int
	attach    map[string]bool
	fromGroup int
	fromUser  int
	fromChat  int
	fromId    map[int]bool
	peerId    map[int]bool
	core      *BotCore

	text       string
	strictMode bool
	word       []string
}

func GetDefaultFilter(core *BotCore) *MessageFilter {
	return &MessageFilter{
		core:      core,
		fromGroup: IFalse,
	}
}

func (filter *MessageFilter) Do(callback MessageListenerCallback) {

	if !filter.strictMode {
		filter.text = strings.ToLower(filter.text)
		if filter.word != nil {
			m := make([]string, 0, len(filter.word))
			for _, w := range filter.word {
				m = append(m, strings.ToLower(w))
			}
			filter.word = m
		}
	}

	filter.core.listeners = append(filter.core.listeners, MessageListener{filter: filter, callback: callback})
}

func (filter *MessageFilter) WithText(text string) *MessageFilter {
	filter.text = strings.TrimSpace(text)
	return filter
}

func (filter *MessageFilter) StrictMode() *MessageFilter {
	filter.strictMode = true
	return filter
}

func (filter *MessageFilter) WithWords(words []string) *MessageFilter {
	filter.word = words
	return filter
}

func (filter *MessageFilter) WithMention() *MessageFilter {
	filter.mention = ITrue
	return filter
}

func (filter *MessageFilter) MoMention() *MessageFilter {
	filter.mention = IFalse
	return filter
}

func (filter *MessageFilter) FromChat() *MessageFilter {
	filter.fromChat = ITrue
	return filter
}

func (filter *MessageFilter) NoChat() *MessageFilter {
	filter.fromChat = IFalse
	return filter
}

func (filter *MessageFilter) FromUser() *MessageFilter {
	filter.fromUser = ITrue
	return filter
}

func (filter *MessageFilter) NoUser() *MessageFilter {
	filter.fromUser = ITrue
	return filter
}

func (filter *MessageFilter) FromGroup() *MessageFilter {
	filter.fromGroup = ITrue
	return filter
}

func (filter *MessageFilter) AnyGroup() *MessageFilter {
	filter.fromGroup = IAny
	return filter
}

func (filter *MessageFilter) WithAttach(attachType string) *MessageFilter {
	filter.attach = make(map[string]bool)
	filter.attach[attachType] = true
	return filter
}

func (filter *MessageFilter) WithAnyAttach(attachTypes []string) *MessageFilter {
	filter.attach = make(map[string]bool)
	for _, attachType := range attachTypes {
		filter.attach[attachType] = true
	}
	return filter
}

func (filter *MessageFilter) FromIds(ids []int) *MessageFilter {
	filter.fromId = make(map[int]bool)
	for _, id := range ids {
		filter.fromId[id] = true
	}
	return filter
}

func (filter *MessageFilter) FromPeerIds(ids []int) *MessageFilter {
	filter.peerId = make(map[int]bool)
	for _, id := range ids {
		filter.peerId[id] = true
	}
	return filter
}

func (filter *MessageFilter) NoGroup() *MessageFilter {
	filter.fromGroup = IFalse
	return filter
}

func compareBool(excepted int, value bool) bool {
	if excepted == IAny {
		return true
	}

	if excepted == IFalse && value == false {
		return true
	}

	if excepted == ITrue && value == true {
		return true
	}

	return false
}

func mapIntersectString(m map[string]bool, array2 []string) bool {
	for _, value := range array2 {
		if _, has := m[value]; has {
			return true
		}
	}

	return false
}

func (filter *MessageFilter) Pass(message *VkApi.CallbackMessage) bool {

	if !compareBool(filter.mention, message.HasMention()) {
		return false
	}

	if !compareBool(filter.fromGroup, message.FromGroup()) {
		return false
	}

	if !compareBool(filter.fromUser, message.FromUser()) {
		return false
	}

	if !compareBool(filter.fromChat, message.FromChat()) {
		return false
	}

	if filter.attach != nil {
		if !mapIntersectString(filter.attach, message.GetAttachTypes()) {
			return false
		}
	}

	if filter.fromId != nil {
		if _, has := filter.fromId[message.FromId]; has == false {
			return false
		}
	}

	if filter.peerId != nil {
		if _, has := filter.peerId[message.PeerId]; has == false {
			return false
		}
	}

	if filter.text != "" {
		var m string

		if message.HasMention() {
			m = message.GetTextWithoutMention()
		} else {
			m = message.Text
		}

		if filter.strictMode == false {
			m = strings.ToLower(m)
		}
		if filter.text != m {
			return false
		}
	}

	if filter.word != nil {
		var m string

		if message.HasMention() {
			m = message.GetTextWithoutMention()
		} else {
			m = message.Text
		}

		if filter.strictMode == false {
			m = strings.ToLower(m)
		}

		for _, w := range filter.word {
			if strings.Index(m, w) != -1 {
				return true
			}
		}
		return false
	}

	return true
}
