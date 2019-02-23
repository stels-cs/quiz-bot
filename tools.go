package main

import (
	"encoding/json"
	"github.com/boltdb/bolt"
	"github.com/stels-cs/vk-api-tools"
	"os"
	"strconv"
	"strings"
)

var defaultLongPollSettings = VkApi.P{
	"api_version":            "5.98",
	"message_new":            "1",
	"message_reply":          "0",
	"photo_new":              "0",
	"audio_new":              "0",
	"video_new":              "0",
	"wall_reply_new":         "0",
	"wall_reply_edit":        "0",
	"wall_reply_delete":      "0",
	"wall_reply_restore":     "0",
	"wall_post_new":          "0",
	"board_post_new":         "0",
	"board_post_edit":        "0",
	"board_post_restore":     "0",
	"board_post_delete":      "0",
	"photo_comment_new":      "0",
	"photo_comment_edit":     "0",
	"photo_comment_delete":   "0",
	"photo_comment_restore":  "0",
	"video_comment_new":      "0",
	"video_comment_edit":     "0",
	"video_comment_delete":   "0",
	"video_comment_restore":  "0",
	"market_comment_new":     "0",
	"market_comment_edit":    "0",
	"market_comment_delete":  "0",
	"market_comment_restore": "0",
	"poll_vote_new":          "0",
	"group_join":             "0",
	"group_leave":            "0",
	"group_change_settings":  "0",
	"group_change_photo":     "0",
	"group_officers_edit":    "0",
	"message_allow":          "0",
	"message_deny":           "0",
	"wall_repost":            "0",
	"user_block":             "0",
	"user_unblock":           "0",
	"messages_edit":          "0",
	"message_typing_state":   "0",
}

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

func trimAndLower(str string) string {
	return strings.TrimSpace(strings.ToLower(str))
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

func GetInt(b *bolt.Bucket, key string) (int, error) {
	val := b.Get([]byte(key))
	if val != nil {
		value, err := strconv.Atoi(string(val))
		if err != nil {
			return 0, err
		}
		return value, nil
	} else {
		return 0, nil
	}
}

func PutInt(b *bolt.Bucket, key string, value int) error {
	return b.Put([]byte(key), []byte(strconv.Itoa(value)))
}

func CreateBucked(db *bolt.DB, name string) error {
	return db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(name))
		if err != nil {
			return err
		}
		return nil
	})
}

func GetIntFromBucked(db *bolt.DB, bucked, key string) (int, error) {
	var i int
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucked))
		var err error
		i, err = GetInt(b, key)
		return err
	})
	return i, err
}

func IncIntFromBucked(db *bolt.DB, bucked, key string) (int, error) {
	var i int
	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucked))
		var err error
		i, err = GetInt(b, key)
		if err != nil {
			return err
		}
		i++
		return PutInt(b, key, i)
	})
	return i, err
}

func PutIntFromBucked(db *bolt.DB, bucked, key string, value int) error {
	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucked))
		return PutInt(b, key, value)
	})
	return err
}

func FillStructureFromBucked(db *bolt.DB, bucked, key string, i interface{}) error {
	return db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucked))
		userRaw := b.Get([]byte(key))
		if userRaw != nil {
			err := json.Unmarshal(userRaw, i)
			return err
		} else {
			return nil
		}
	})
}

func PutStructure(b *bolt.Bucket, key string, value interface{}) error {
	raw, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return b.Put([]byte(key), raw)
}

func PutStructureIntoBucked(db *bolt.DB, bucked, key string, value interface{}) error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucked))
		return PutStructure(b, key, value)
	})
}

func DeleteBucked(db *bolt.DB, name string) error {
	return db.Update(func(tx *bolt.Tx) error {
		return tx.DeleteBucket([]byte(name))
	})
}
