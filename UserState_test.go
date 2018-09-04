package main

import (
	"encoding/json"
	"math/rand"
	"sort"
	"testing"
	"time"
)

func TestUnicodeMarshal(t *testing.T) {

	str := "HELLO & WORLD!"
	bstr, err := json.Marshal(str)
	if err != nil {
		t.Error(err)
	}
	println(string(bstr))
}

func TestUserStateSimple(t *testing.T) {
	state, err := GetUserState(false, "")
	if err != nil {
		t.Error(err)
	}

	state.AddScore(100)
	state.AddScore(100)
	state.AddScore(100)

	state.AddScore(1000)
	state.AddScore(1000)
	state.AddScore(1000)
	state.AddScore(1000)

	state.AddScore(10000)
	state.AddScore(10000)
	state.AddScore(10000)
	state.AddScore(10000)
	state.AddScore(10000)

	me, top := state.GetTop(1000, 10)

	if me == nil {
		t.Error("me is nil, must be *UserItem")
	}

	if me.GetPlace() != 2 {
		t.Error("Expected me.GetPlace() = 2, got", me.GetPlace())
	}

	if me.GetScore() != 4 {
		t.Error("Exprected me.GetScore = 4, got", me.GetScore())
	}

	if len(top) != 3 {
		t.Error("Expected let(top) = 3, got", len(top))
	}

	trueTop := []int{10000, 1000, 100}
	for place, id := range trueTop {
		if top[place].GetPlace() != place+1 {
			t.Errorf("Expected top[%d].GetPlace() = %d, got %d", place, place+1, top[place].GetPlace())
		}
		if top[place].Id != id {
			t.Errorf("Expected top[%d].Id = %d, got %d", place, id, top[place].Id)
		}
		if top[place].GetScore() != 5-place {
			t.Errorf("Expected top[%d].GetScore = %d, got %d", place, 5-place, top[place].GetScore())
		}
	}
}

func TestUserStateHard(t *testing.T) {
	state, err := GetUserState(false, "")
	if err != nil {
		t.Error(err)
	}

	rand.Seed(time.Now().UnixNano())

	top := make([]*UserItem, 0)

	y := rand.Int() % 100
	for i := 0; i < 100; i++ {
		u := &UserItem{
			ScoreValue: i*y + i + i*2 + i*3 + i + 2,
			Id:         i + 1000,
		}
		u.PlaceValue = u.GetScore()
		top = append(top, u)
	}

	for _, u := range top {
		state.GetUserSafety(u.Id)
	}

	//state.PrintTree(0)

	top10 := RR(append([]*UserItem(nil), top...))
	sort.Sort(top10)

	top10 = top10[:10]

	for len(top) > 0 {
		index := rand.Int() % len(top)
		u := top[index]

		state.AddScore(u.Id)
		u.ScoreValue--

		if u.GetScore() <= 0 {
			top = append(top[:index], top[index+1:]...)
		}
	}

	me, checkTop10 := state.GetTop(top10[0].Id, 10)

	if me.Id != top10[0].Id {
		t.Error("1")
	}

	if me.GetScore() != top10[0].GetPlace() {
		t.Error("2")
	}

	if me.GetPlace() != 1 {
		t.Error("Me place is not 1, got", me.GetPlace(), " me")
	}

	if len(checkTop10) != len(top10) {
		t.Error("4")
	}

	for index, u := range checkTop10 {

		if u.GetPlace() != index+1 {
			t.Error("Bad place number, expected", index+1, "got", u.GetPlace())
		}

		if u.GetScore() != top10[index].GetPlace() {
			t.Error("Bad score expected", top10[index].GetPlace(), "got", u.GetScore())
		}

		if u.Id != top10[index].Id {
			t.Error("Bad id, expeced", top10[index].Id, "got", u.Id)
		}

		if u != state.GetUserSafety(u.Id) {
			t.Error("8")
		}
	}
}

type RR []*UserItem

func (r RR) Len() int {
	return len(r)
}

func (r RR) Less(i int, j int) bool {
	ui := (r)[i]
	uj := (r)[j]

	return ui.GetScore() > uj.GetScore()
}

func (r RR) Swap(i int, j int) {
	ui := (r)[i]
	uj := (r)[j]

	(r)[i] = uj
	(r)[j] = ui
}

func TestGetTopOnEmptyObject(t *testing.T) {
	state, err := GetUserState(false, "")
	if err != nil {
		t.Error(err)
	}

	me, top := state.GetTop(10, 10)

	if me.GetPlace() != 1 {
		t.Error("Bad place, expected 1, got", me.GetPlace())
	}

	if len(top) != 1 {
		t.Error("Bad top count, expected 1, got", len(top))
	}

	me, top = state.GetTop(20, 10)

	if me.GetPlace() != 2 {
		t.Error("Bad place, expected 2, got", me.GetPlace())
	}

	if len(top) != 2 {
		t.Error("Bad top count, expected 2, got", len(top))
	}

	me, top = state.GetTop(10, 10)

	if me.GetPlace() != 1 {
		t.Error("Bad place, expected 1, got", me.GetPlace())
	}

	if len(top) != 2 {
		t.Error("Bad top count, expected 2, got", len(top))
	}

	me, top = state.GetTop(10, 10)

	if me.GetPlace() != 1 {
		t.Error("Bad place, expected 1, got", me.GetPlace())
	}

	if len(top) != 2 {
		t.Error("Bad top count, expected 2, got", len(top))
	}
}

func TestSateWithPersistentDb(t *testing.T) {
	state, err := GetUserState(true, "test_db.db")
	if err != nil {
		t.Error(err)
	}

	state.SaveTime = 1

	state.DropTop()

	state.GetUserSafety(10)
	state.GetUserSafety(20)
	state.GetUserSafety(30)

	time.Sleep(2 * time.Second)
	state.Save()
	state.Clear()

	state.GetUserSafety(40)

	time.Sleep(2 * time.Second)
	state.Save()
	state.Clear()

	state.GetUserSafety(30)

	_, top := state.GetTop(10, 10)

	if len(top) != 4 {
		t.Error("Expected top len 4 got", len(top))
	}
}
