package main

import (
	"bufio"
	"math/rand"
	"os"
	"strings"
)

type Question struct {
	Text   string
	Answer string
}

type QuestionPoll struct {
	List []Question
}

func (p *QuestionPoll) GetQuestion() *Question {
	if len(p.List) > 0 {
		return &p.List[rand.Intn(len(p.List))]
	} else {
		return &Question{"No question was loaded", "fuck"}
	}
}

func (p *QuestionPoll) LoadFromFile(fName string) error {
	file, err := os.Open(fName)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		str := strings.Split(scanner.Text(), "|")
		if len(str) == 2 {
			p.List = append(p.List, Question{str[0], trimAndLower(str[1])})
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

func (q *Question) GetBuzzyText() string {
	n := rand.Intn(101)
	if n > 0 && n < 33 {
		return q.Text
	} else if n >= 33 && n < 55 {
		return strings.Replace(q.Text, "е", "e", 4)
	} else if n >= 55 && n < 66 {
		return strings.Replace(q.Text, "а", "a", 4)
	} else {
		return strings.Replace(q.Text, "о", "o", 4)
	}

}
