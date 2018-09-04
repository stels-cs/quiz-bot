package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/boltdb/bolt"
	"strconv"
	"strings"
	"sync"
	"time"
)

const TopBucked = "TopBucked"

const HeadKey = "head"
const TailKey = "tail"

type UserState struct {
	index      map[int]*UserItem
	head       int
	tail       int
	db         *bolt.DB
	Persistent bool
	SaveTime   int
	createLock *sync.Mutex
}

func GetUserState(Persistent bool, databaseName string) (*UserState, error) {
	x := &UserState{
		index:      make(map[int]*UserItem),
		head:       0,
		tail:       0,
		Persistent: Persistent,
		SaveTime:   20,
		createLock: &sync.Mutex{},
	}

	if Persistent {
		db, err := bolt.Open(databaseName, 0600, &bolt.Options{Timeout: 1 * time.Second})
		if err != nil {
			return nil, err
		}
		x.db = db
		CreateBucked(db, TopBucked)
		if x.head, err = GetIntFromBucked(db, TopBucked, HeadKey); err != nil {
			return nil, err
		}
		if x.tail, err = GetIntFromBucked(db, TopBucked, TailKey); err != nil {
			return nil, err
		}
	}
	return x, nil
}

func (s *UserState) AddScoreValue(userId, value int) bool {
	s.createLock.Lock()
	defer s.createLock.Unlock()
	item := s.get(userId)
	item.AddScoreValue(value)
	return s.upItem(item)
}

func (s *UserState) AddScore(userId int) bool {
	s.createLock.Lock()
	defer s.createLock.Unlock()

	item := s.get(userId)
	item.AddScore()

	return s.upItem(item)
}

func (s *UserState) GetTop(userId int, size int) (*UserItem, []*UserItem) {
	s.createLock.Lock()
	defer s.createLock.Unlock()

	user := s.get(userId)
	top := make([]*UserItem, 0, size)
	h := s.head
	for len(top) < size && h != 0 {
		user := s.get(h)
		top = append(top, user)
		h = user.GetNext()
	}

	return user, top
}

func (s *UserState) get(userId int) *UserItem {
	i := s.index[userId]
	if i != nil {
		return i
	} else {
		i = s.createItem(userId)
		return i
	}
}

func (s *UserState) GetUserSafety(userId int) *UserItem {
	s.createLock.Lock()
	defer s.createLock.Unlock()
	return s.get(userId)
}

func (s *UserState) createItem(userId int) *UserItem {
	i := GetDefaultUserItem(userId)

	if s.Persistent {
		if err := FillStructureFromBucked(s.db, TopBucked, s.getKey(userId), i); err != nil {
			panic(err)
		}
	}

	if i.IsNew() {
		i.ClearNew()
		s.pushItem(i)
	} else {
		i.ClearNew()
		s.index[i.Id] = i
	}
	return i
}

func (s *UserState) pushItem(item *UserItem) {
	s.index[item.Id] = item

	if s.head == 0 || s.head == item.Id {
		s.head = item.Id
		s.tail = item.Id
	} else if s.tail == 0 || s.tail == item.Id {
		item.SetPlace(s.get(s.head).GetPlace() + 1)
		s.tail = item.Id
		s.get(s.head).SetNext(item.Id)
		s.get(s.tail).SetBefore(s.head)
	} else {
		item.SetPlace(s.get(s.tail).GetPlace() + 1)
		item.SetBefore(s.tail)
		s.get(s.tail).SetNext(item.Id)
		s.tail = item.Id
		s.upItem(item)
	}
}

func (s *UserState) upItem(item *UserItem) bool {
	if item.GetBefore() != 0 && s.get(item.GetBefore()).GetScore() < item.GetScore() {
		before := s.get(item.GetBefore())

		if before.GetBefore() != 0 {
			s.get(before.GetBefore()).SetNext(item.Id)
		} else {
			s.head = item.Id
		}

		if item.GetNext() != 0 {
			s.get(item.GetNext()).SetBefore(before.Id)
		} else {
			s.tail = before.Id
		}

		item.Swap(before)
		s.upItem(item)
		return true
	}
	return false
}

func (s *UserState) getKey(id int) string {
	return "U_" + strconv.Itoa(id)
}

func (s *UserState) Save() {
	s.createLock.Lock()
	defer s.createLock.Unlock()

	index := make(map[int]*UserItem)
	lim := time.Now().Unix() - int64(s.SaveTime)
	for userId, data := range s.index {
		if data.GetTime() > 0 && data.GetTime() < lim {
			data.Clear()
			index[userId] = data
		}
	}
	s.saveData(index)
}

func (s *UserState) Clear() {
	s.createLock.Lock()
	defer s.createLock.Unlock()

	ids := make([]int, 0, 100)
	for userId, data := range s.index {
		if s.tail != userId && s.head != userId {
			if data.IsDirty() == false {
				ids = append(ids, userId)
			}
		}
	}
	for _, id := range ids {
		delete(s.index, id)
	}
}

func (s *UserState) SaveForce() {
	s.createLock.Lock()
	defer s.createLock.Unlock()
	s.saveData(s.index)
}

func (s *UserState) saveData(index map[int]*UserItem) {
	if !s.Persistent {
		return
	}

	err := s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(TopBucked))
		for userId, data := range index {
			if err := PutStructure(b, s.getKey(userId), data); err != nil {
				return err
			}
		}

		if err := PutInt(b, HeadKey, s.head); err != nil {
			return err
		}

		if err := PutInt(b, TailKey, s.tail); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		panic(err)
	}
}

func (s *UserState) DropTop() {
	s.createLock.Lock()
	defer s.createLock.Unlock()
	s.head = 0
	s.tail = 0
	s.index = make(map[int]*UserItem)
	if s.Persistent {
		if err := DeleteBucked(s.db, TopBucked); err != nil {
			panic(err)
		}
		if err := CreateBucked(s.db, TopBucked); err != nil {
			panic(err)
		}
	}
}

func (s *UserState) Unroll(userId int) {
	s.createLock.Lock()
	defer s.createLock.Unlock()
	u := s.get(userId)
	if u.GetNext() != 0 {
		n := s.get(u.GetNext())
		u.SetScore(n.GetScore() + 1)
	}
}

func (s *UserState) RebuildTree() {
	if !s.Persistent {
		return
	}

	s.SaveForce()

	s.createLock.Lock()
	s.head = 0
	s.tail = 0
	s.index = make(map[int]*UserItem)

	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(TopBucked))

		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			fmt.Printf("key=%s, value=%s\n", k, v)
			if strings.Index(string(k), "U") == 0 {
				u := GetDefaultUserItem(0)
				err := json.Unmarshal(v, u)
				if err != nil {
					return err
				}
				if u.Id == 0 {
					return errors.New("User id is ZERO!!!!! ")
				}

				u.SetPlace(len(s.index) + 1)
				u.SetNext(0)
				u.SetBefore(0)

				s.pushItem(u)
			}
		}

		return nil
	})
	s.createLock.Unlock()

	s.SaveForce()

	if err != nil {
		panic(err)
	}
}
