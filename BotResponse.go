package main

type BotMessage struct {
	Message         string
	RandomId        int
	Lat             float64
	Long            float64
	Attachment      []string
	ForwardMessages []int
	StickerId       int
	Keyboard        *Keyboard
}

func TextMessage(text string) *BotMessage {
	return &BotMessage{
		Message: text,
	}
}

type BotResponse struct {
	Message   *BotMessage
	Typing    bool
	Read      bool
	Answered  bool
	Important bool
}

func NoReactionResponse() *BotResponse {
	return &BotResponse{}
}

func SimpleMessageResponse(text string) *BotResponse {
	return &BotResponse{
		Message: TextMessage(text),
	}
}

func (msg *BotMessage) SetKeyboard(kbd *Keyboard) *BotMessage {
	msg.Keyboard = kbd
	return msg
}

func (res *BotResponse) SetKeyboard(kbd *Keyboard) *BotResponse {
	if res.Message != nil {
		res.Message.SetKeyboard(kbd)
	}
	return res
}
