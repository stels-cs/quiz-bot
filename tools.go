package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"
)

func ff(cond bool, t string, f string) string {
	if cond {
		return t
	} else {
		return f
	}
}

func env(name string, def string) string {
	v := os.Getenv(name)
	return ff(v != "", v, def)
}

func transChoose(x int, one string, two string, five string) string {
	if x == 0 {
		return five
	}
	if x > 20 || x < 10 {
		x = x % 10
		if x == 1 {
			return one
		} else if x >= 2 && x <= 4 {
			return two
		} else {
			return five
		}
	} else {
		return five
	}
}

func inArray(haystack []string, needle string) bool {
	for _, v := range haystack {
		if v == needle {
			return true
		}
	}
	return false
}

func trimAndLower(str string) string {
	return strings.TrimSpace(strings.ToLower(str))
}

func postJsonRequest(url string, request string, v interface{}) ([]byte, error) {
	t := &http.Client{Timeout: time.Second * 300}
	var r []byte
	resp, err := t.Post(url, "application/json", bytes.NewBuffer([]byte(request)))
	if err != nil {
		return r, err
	}
	data, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return r, err
	}
	err = json.Unmarshal(data, v)
	if err != nil {
		return r, err
	} else {
		return data, nil
	}
}

func trimToken(token string) string {
	if len(token) > 10 {
		left := string([]rune(token)[0:5])
		right := string([]rune(token)[len(token)-6 : len(token)])
		return left + "..." + right
	} else {
		return token
	}
}

// DistanceForStrings returns the edit distance between source and target.
//
// It has a runtime proportional to len(source) * len(target) and memory use
// proportional to len(target).
func DistanceForStrings(source []rune, target []rune) int {
	// Note: This algorithm is a specialization of MatrixForStrings.
	// MatrixForStrings returns the full edit matrix. However, we only need a
	// single value (see DistanceForMatrix) and the main loop of the algorithm
	// only uses the current and previous row. As such we create a 2D matrix,
	// but with height 2 (enough to store current and previous row).
	height := len(source) + 1
	width := len(target) + 1
	matrix := make([][]int, 2)

	// Initialize trivial distances (from/to empty string). That is, fill
	// the left column and the top row with row/column indices.
	for i := 0; i < 2; i++ {
		matrix[i] = make([]int, width)
		matrix[i][0] = i
	}
	for j := 1; j < width; j++ {
		matrix[0][j] = j
	}

	// Fill in the remaining cells: for each prefix pair, choose the
	// (edit history, operation) pair with the lowest cost.
	for i := 1; i < height; i++ {
		cur := matrix[i%2]
		prev := matrix[(i-1)%2]
		cur[0] = i
		for j := 1; j < width; j++ {
			delCost := prev[j] + 2
			matchSubCost := prev[j-1]
			if source[i-1] != target[j-1] {
				matchSubCost += 1
			}
			insCost := cur[j-1] + 2
			cur[j] = min(delCost, min(matchSubCost, insCost))
		}
	}
	return matrix[(height-1)%2][width-1]
}

func min(a int, b int) int {
	if b < a {
		return b
	}
	return a
}
