package main

import (
	"encoding/json"
)

type ButtonAction struct {
	Type    string `json:"type"`
	Payload string `json:"payload"`
	Label   string `json:"label"`
}

type Button struct {
	Action ButtonAction `json:"action"`
	Color  string       `json:"color"`
}

func GetBtn(label string, payload string, color string) (*Button, error) {
	p, err := json.Marshal(interface{}(payload))
	if err != nil {
		return nil, err
	}
	ba := ButtonAction{
		Type:    "text",
		Payload: string(p),
		Label:   label,
	}
	return &Button{
		Action: ba,
		Color:  color,
	}, nil
}

func GetDefault(label string, payload string) (*Button, error) {
	return GetBtn(label, payload, "default")
}

func GetNegative(label string, payload string) (*Button, error) {
	return GetBtn(label, payload, "negative")
}

func GetPositive(label string, payload string) (*Button, error) {
	return GetBtn(label, payload, "positive")
}

func GetPrimary(label string, payload string) (*Button, error) {
	return GetBtn(label, payload, "primary")
}

type Keyboard struct {
	OneTime bool        `json:"one_time"`
	Buttons [][]*Button `json:"buttons"`
}

func GetDefaultkeyboad() *Keyboard {
	kbd, err := GetDefaultKbd()
	if err != nil {
		return nil
	} else {
		return kbd
	}
}

func GetDefaultKbd() (*Keyboard, error) {
	start, err := GetPrimary("Начать игру", startGameCommand)
	if err != nil {
		return nil, err
	}
	top, err := GetDefault("Рейтинг", topCommand)
	if err != nil {
		return nil, err
	}

	help, err := GetDefault("Помощь", helpCommand)
	if err != nil {
		return nil, err
	}

	btns := [][]*Button{
		{
			start,
		},
		{
			top, help,
		},
	}
	kbd := Keyboard{
		true,
		btns,
	}
	return &kbd, nil
}

func GetStopKeyboad() *Keyboard {
	kbd, err := GetStopKbd()
	if err != nil {
		return nil
	} else {
		return kbd
	}
}

func GetStopKbd() (*Keyboard, error) {
	stop, err := GetDefault("Закончить игру", stopGameCommand)
	if err != nil {
		return nil, err
	}
	btns := [][]*Button{
		{
			stop,
		},
	}
	kbd := Keyboard{
		false,
		btns,
	}
	return &kbd, nil
}
