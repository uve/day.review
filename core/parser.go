package core

import (
	"appengine"
	"appengine/urlfetch"
	"fmt"
	"io/ioutil"
	//"net/http"

	"encoding/json"
	"errors"
	"time"
)

func parser(c appengine.Context) (*Account, error) {
	accounts := make([]*Account, 0)

	parsed := 0
	var value *Account
	ch := make(chan *Account) // We don't need any data to be passed, so use an empty struct

	for _, follower := range followers {
		go getAccount(c, ch, follower)
	}

	// Since we spawned 100 routines, receive 100 messages.
	for _, follower := range followers {
		value = <-ch
		accounts = append(accounts, value)
		parsed = parsed + 1
		c.Debugf("parsed: ", follower, "\ttotal: ", parsed, "\tLikes: ", value.Likes)
	}

	account := filterAccountMaxLikes(accounts)

	c.Debugf(fmt.Sprintf("Result: https://www.instagram.com/p/%v", account.Node.Code))
	c.Debugf(fmt.Sprintf("nUser: ", account.UserID))
	c.Debugf(fmt.Sprintf("nLikes: ", account.Likes))

	return account, nil
}

func getAccount(c appengine.Context, ch chan<- *Account, userID string) {
	account, err := parserAccount(c, userID)
	if err == nil {
		ch <- account
	} else {
		c.Debugf("Error: ", userID, err)
		ch <- &Account{}
	}
}

func parserAccount(c appengine.Context, userID string) (*Account, error) {
	c.Debugf("Parsing: ", userID)

	user, err := parserUser(c, userID)
	if err != nil {
		return nil, err
	}

	c.Debugf(userID, " Total media:", len(user.Media.Nodes))
	nodes := filterMediaByData(user.Media.Nodes)
	nodes = filterNotVideo(nodes)

	if len(nodes) < 1 {
		return nil, errors.New("No media found")
	}

	c.Debugf(userID, " For 1 day: ", len(nodes))

	node := filterMaxLikes(nodes)

	c.Debugf(userID, " Likes: ", node.Likes.Count)

	return &Account{
		User:   user,
		UserID: userID,
		Node:   node,
		Likes:  node.Likes.Count,
	}, nil

}

func filterMediaByData(nodes []Node) []Node {
	filtered := make([]Node, 0)
	now := time.Now()
	dayAgo := now.AddDate(0, 0, -1).Unix()

	for _, node := range nodes {
		if node.Date > dayAgo {
			filtered = append(filtered, node)
		}
	}

	return filtered
}

func filterNotVideo(nodes []Node) []Node {
	filtered := make([]Node, 0)

	for _, node := range nodes {
		if node.IsVideo == false {
			filtered = append(filtered, node)
		}
	}

	return filtered
}

func parserUser(c appengine.Context, id string) (*User, error) {
	src := "https://www.instagram.com/" + id + "/?__a=1"
	//c := appengine.NewContext(r)
	c.Debugf("parsing: ", id)

	client := urlfetch.Client(c)
	resp, err := client.Get(src)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var userJSON *UserJSON
	err = json.Unmarshal(body, &userJSON)
	if err != nil {
		return nil, err
	}

	return &(userJSON.User), nil
}

func filterMaxLikes(nodes []Node) Node {
	var filtered Node
	count := 0

	for _, node := range nodes {
		if node.Likes.Count > count {
			filtered = node
			count = node.Likes.Count
		}
	}

	return filtered
}

func filterAccountMaxLikes(nodes []*Account) *Account {
	var filtered *Account
	count := 0

	for _, node := range nodes {
		if node.Likes > count {
			filtered = node
			count = node.Likes
		}
	}

	return filtered
}
