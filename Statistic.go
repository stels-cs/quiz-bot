package main

import (
	"fmt"
	"github.com/boltdb/bolt"
	"sync"
	"time"
)

const StatisticBucked = "StatisticBucked"

const ReceiveMessages = "RM"
const SendMessage = "SM"
const StartGames = "SG"
const StartQuestions = "SQ"
const DoneQuestions = "DQ"
const DialogsCount = "DIQ"
const UserInTopCount = "UIT"

type Statistic struct {
	db   *bolt.DB
	lock *sync.Mutex
}

func GetStatistic(databaseName string) (*Statistic, error) {
	x := &Statistic{
		lock: &sync.Mutex{},
	}

	db, err := bolt.Open(databaseName, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, err
	}
	x.db = db
	CreateBucked(db, StatisticBucked)
	return x, nil
}

func (statistic *Statistic) GetKeyName(prefix string) string {
	now := time.Now()

	return statistic.GetKeyNameAt(prefix, now)
}

func (statistic *Statistic) GetKeyNameAt(prefix string, now time.Time) string {
	return fmt.Sprintf("%s_%d_%d_%d", prefix, now.Day(), now.Month(), now.Year())
}

func (statistic *Statistic) ReceiveMessage() {
	statistic.lock.Lock()
	defer statistic.lock.Unlock()

	IncIntFromBucked(statistic.db, StatisticBucked, statistic.GetKeyName(ReceiveMessages))
}

func (statistic *Statistic) SendMessage() {
	statistic.lock.Lock()
	defer statistic.lock.Unlock()

	IncIntFromBucked(statistic.db, StatisticBucked, statistic.GetKeyName(SendMessage))
}

func (statistic *Statistic) StartGame() {
	statistic.lock.Lock()
	defer statistic.lock.Unlock()

	IncIntFromBucked(statistic.db, StatisticBucked, statistic.GetKeyName(StartGames))
}

func (statistic *Statistic) StartQuestion() {
	statistic.lock.Lock()
	defer statistic.lock.Unlock()

	IncIntFromBucked(statistic.db, StatisticBucked, statistic.GetKeyName(StartQuestions))
}

func (statistic *Statistic) DoneQuestions() {
	statistic.lock.Lock()
	defer statistic.lock.Unlock()

	IncIntFromBucked(statistic.db, StatisticBucked, statistic.GetKeyName(DoneQuestions))
}

func (statistic *Statistic) PutDialogCount(count int) {
	statistic.lock.Lock()
	defer statistic.lock.Unlock()

	max, _ := GetIntFromBucked(statistic.db, StatisticBucked, DialogsCount)

	if count > max {
		PutIntFromBucked(statistic.db, StatisticBucked, statistic.GetKeyName(DialogsCount), count)
		PutIntFromBucked(statistic.db, StatisticBucked, DialogsCount, count)
	}
}

func (statistic *Statistic) PutUsersInTop(count int) {
	statistic.lock.Lock()
	defer statistic.lock.Unlock()

	PutIntFromBucked(statistic.db, StatisticBucked, statistic.GetKeyName(UserInTopCount), count)
}

func (statistic *Statistic) GetTopLine(prefix string) []int {
	statistic.lock.Lock()
	defer statistic.lock.Unlock()

	now := time.Now()

	now = now.Add(7 * time.Hour * 24 * -1)

	values := make([]int, 0, 7)

	for i := 0; i <= 7; i++ {
		i, _ := GetIntFromBucked(statistic.db, StatisticBucked, statistic.GetKeyNameAt(prefix, now))
		values = append(values, i)
		now = now.Add(time.Hour * 24)
	}

	return values
}

func (statistic *Statistic) SetTsValue(value int) {
	statistic.lock.Lock()
	defer statistic.lock.Unlock()

	PutIntFromBucked(statistic.db, StatisticBucked, "TS", value)
}

func (statistic *Statistic) GetTsValue() (int, error) {
	statistic.lock.Lock()
	defer statistic.lock.Unlock()

	return GetIntFromBucked(statistic.db, StatisticBucked, "TS")
}
