package main

import (
	"time"
)

type UserItem struct {
	Id          int `json:"id"`
	BeforeValue int `json:"before"`
	NextValue   int `json:"next"`
	PlaceValue  int `json:"place"`
	ScoreValue  int `json:"score"`
	PeerId      int `json:"peer_id"`
	dirtyTime   int64
}

func (u *UserItem) SetNext(v int) {
	u.NextValue = v
	u.UpTime()
}

func (u *UserItem) SetBefore(v int) {
	u.BeforeValue = v
	u.UpTime()
}

func (u *UserItem) SetScore(v int) {
	u.ScoreValue = v
	u.UpTime()
}

func (u *UserItem) SetPlace(v int) {
	u.PlaceValue = v
	u.UpTime()
}

func (u *UserItem) AddScore() {
	u.ScoreValue++
	u.UpTime()
}

func (u *UserItem) AddScoreValue(value int) {
	u.ScoreValue += value
	u.UpTime()
}

func (u *UserItem) GetPlace() int {
	return u.PlaceValue
}

func (u *UserItem) GetNext() int {
	return u.NextValue
}

func (u *UserItem) GetBefore() int {
	return u.BeforeValue
}

func (u *UserItem) GetScore() int {
	return u.ScoreValue
}

func (u *UserItem) Swap(before *UserItem) {
	u.PlaceValue, before.PlaceValue = before.PlaceValue, u.PlaceValue
	u.BeforeValue, before.BeforeValue = before.BeforeValue, u.Id
	before.NextValue, u.NextValue = u.NextValue, before.Id
	before.UpTime()
	u.UpTime()
}

func (u *UserItem) GetTime() int64 {
	return u.dirtyTime
}

func (u *UserItem) Clear() {
	u.dirtyTime = 0
}

func (u *UserItem) UpTime() {
	u.dirtyTime = time.Now().Unix()
}

func (u *UserItem) IsDirty() bool {
	return u.dirtyTime != 0
}

func (u *UserItem) GetGroupId() int {
	return u.PeerId
}

func (u *UserItem) SetGroupId(id int) {
	u.PeerId = id
	u.UpTime()
}

func (u *UserItem) IsNew() bool {
	return u.ScoreValue == -1
}

func (u *UserItem) ClearNew() {
	if u.IsNew() {
		u.ScoreValue = 0
	}
}

func GetDefaultUserItem(userId int) *UserItem {
	return &UserItem{
		Id:          userId,
		ScoreValue:  -1,
		PlaceValue:  1,
		BeforeValue: 0,
		NextValue:   0,
		dirtyTime:   0,
	}
}
