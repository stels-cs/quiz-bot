package main

import (
	"github.com/boltdb/bolt"
	"sort"
	"strconv"
	"sync"
	"time"
)

const ConversationBucked = "ConversationBucked"

type ConversationTop struct {
	mutex sync.Mutex
	cache map[int]map[int]int
	db    *bolt.DB
}

func GetConversationTop(databaseName string) (*ConversationTop, error) {

	db, err := bolt.Open(databaseName, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, err
	}
	CreateBucked(db, ConversationBucked)

	return &ConversationTop{
		mutex: sync.Mutex{},
		db:    db,
		cache: make(map[int]map[int]int),
	}, nil
}

func (top *ConversationTop) AddScore(peerId, userId int) error {
	top.mutex.Lock()
	defer top.mutex.Unlock()

	var err error
	if _, has := top.cache[peerId]; has == false {
		top.cache[peerId], err = top.load(peerId)
	}

	top.cache[peerId][userId]++
	return err
}

type SortedTop []*UserItem

func (c SortedTop) Len() int           { return len(c) }
func (c SortedTop) Swap(i, j int)      { c[i], c[j] = c[j], c[i] }
func (c SortedTop) Less(i, j int) bool { return c[i].GetScore() > c[j].GetScore() }

func (top *ConversationTop) GetTop(peerId int) ([]*UserItem, error) {
	top.mutex.Lock()
	defer top.mutex.Unlock()

	var err error
	if _, has := top.cache[peerId]; has == false {
		top.cache[peerId], err = top.load(peerId)
	}

	res := make([]*UserItem, 0, len(top.cache[peerId]))

	for userId, score := range top.cache[peerId] {
		res = append(res, &UserItem{
			Id:         userId,
			ScoreValue: score,
		})
	}

	sort.Sort(SortedTop(res))

	for index, item := range res {
		item.SetPlace(index + 1)
	}

	return res, err
}

func (top *ConversationTop) Remove(peerId int) error {
	top.mutex.Lock()
	defer top.mutex.Unlock()

	if _, has := top.cache[peerId]; has {
		err := top.save(peerId, top.cache[peerId])

		delete(top.cache, peerId)
		return err
	}
	return nil
}

func (top *ConversationTop) SaveAll() error {
	for peerId, t := range top.cache {
		err := top.save(peerId, t)
		if err != nil {
			return err
		}
	}
	return nil
}

func (top *ConversationTop) load(peerId int) (map[int]int, error) {
	t := make(map[int]int)

	err := FillStructureFromBucked(top.db, ConversationBucked, strconv.Itoa(peerId), &t)

	return t, err
}

func (top *ConversationTop) save(peerId int, data map[int]int) error {
	return PutStructureIntoBucked(top.db, ConversationBucked, strconv.Itoa(peerId), data)
}
