package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
)

type Top struct {
	data       map[int]int
	fast       map[int]int
	fName      string
	logger     *log.Logger
	lockSave   bool
	writeMutex *sync.Mutex
	saveMutex  *sync.Mutex
}

func GetTop(fName string, logger *log.Logger) Top {
	return Top{
		map[int]int{},
		map[int]int{},
		fName,
		logger,
		false,
		&sync.Mutex{},
		&sync.Mutex{}}
}

func (top *Top) Load() error {
	file, err := os.Open(top.fName)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		str := strings.Split(scanner.Text(), "|")
		if len(str) == 2 {
			userId, u := strconv.Atoi(strings.TrimSpace(str[0]))
			rating, r := strconv.Atoi(strings.TrimSpace(str[1]))
			if u == nil && r == nil {
				top.data[userId] = rating
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

func (top *Top) Save() error {
	file, err := os.Create(top.fName)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for k, v := range top.data {
		_, err := writer.WriteString(fmt.Sprintf("%d|%d\n", k, v))
		if err != nil {
			return err
		}
	}
	writer.Flush()

	return nil
}

func (top *Top) Inc(userId int) int {
	top.writeMutex.Lock()
	_, ok := top.data[userId]
	if ok {
		top.data[userId]++
	} else {
		top.data[userId] = 1
	}
	top.writeMutex.Unlock()

	if top.lockSave == false {
		go top.SaveWithLock()
	}
	return top.data[userId]
}

func (top *Top) Clear(userId int) {
	top.writeMutex.Lock()
	_, ok := top.data[userId]
	if ok {
		top.data[userId] = 0
	}

	_, fok := top.fast[userId]
	if fok {
		delete(top.fast, userId)
	}

	top.writeMutex.Unlock()
}

func insertIntoRating(arr *[10][2]int, startAt int, rating int, id int) {
	for i := 8; i >= startAt; i-- {
		arr[i+1][0] = arr[i][0]
		arr[i+1][1] = arr[i][1]
	}

	arr[startAt][0] = rating
	arr[startAt][1] = id
}

func (top *Top) GetTop10() [10][2]int {
	list := [10][2]int{
		{0, 0},
		{0, 0},
		{0, 0},
		{0, 0},
		{0, 0},
		{0, 0},
		{0, 0},
		{0, 0},
		{0, 0},
		{0, 0},
	}

	for userId, rating := range top.data {
		if rating > list[9][0] {
			for i := 0; i < 10; i++ {
				if list[i][0] < rating {
					insertIntoRating(&list, i, rating, userId)
					break
				}
			}
		}
	}

	return list
}
func (top *Top) SaveWithLock() {
	top.saveMutex.Lock()
	top.lockSave = true
	err := top.Save()
	if err != nil {
		top.logger.Println(err.Error())
	}
	top.lockSave = false
	top.saveMutex.Unlock()
}

func (top *Top) FastUser(id int) {
	top.writeMutex.Lock()
	_, ok := top.fast[id]
	if ok {
		top.fast[id]++
	} else {
		top.fast[id] = 1
	}
	top.writeMutex.Unlock()
}

func (top *Top) GetFastUsers() string {
	max := 0
	id := 0

	top.writeMutex.Lock()
	for uid, count := range top.fast {
		if count > max {
			max = count
			id = uid
		}
	}
	top.writeMutex.Unlock()

	if max > 0 {
		return fmt.Sprintf(`@id%d -- %d`, id, max)
	} else {
		return ""
	}

}
