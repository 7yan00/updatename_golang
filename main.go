package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/mrjones/oauth"
)

type user struct {
	Id         int64  `json:"id"`
	Name       string `json:"name"`
	ScreenName string `json:"screen_name"`
}
type status struct {
	Id   uint64 `json:"id"`
	Text string `json:"text"`
	User user   `json:"user"`
}

var consumerKey *string = flag.String(
	"consumerkey",
	"mEF22DxPk6cocNoc3lQQBoj55",
	"Consumer Key from Twitter. See: https://dev.twitter.com/apps/new")

var consumerSecret *string = flag.String(
	"consumersecret",
	"cGOq2NGmEqdwzVPPkQfMJuh6HEVFuVz5qFqBQJAteVuKC4ZQS9",
	"Consumer Secret from Twitter. See: https://dev.twitter.com/apps/new")

var accessToken *oauth.AccessToken

func main() {
	ScreenName := "ryusen33"
	re := regexp.MustCompile(`^(.+)\(@` + ScreenName + `\)$`)
	flag.Parse()
	fmt.Println("loading consumerkey......")
	loading()

	get_timeline(func(b []byte) {
		s := new(status)
		err := json.Unmarshal(b, s)
		if err != nil {
			return
		}

		match := re.FindStringSubmatch(s.Text)
		if len(match) != 2 {
			return
		}

		newName := strings.TrimSpace(match[1])
		err = updateName(newName)

		if err != nil {
			fmt.Println("falied")
			fmt.Println(err)

		}

		fmt.Println(newName)
		err = UpdateStatus(fmt.Sprintf("@%v 「%v」に改名したのです", s.User.ScreenName, newName), s.Id)

		if err != nil {
			fmt.Println("tweet failed")
			fmt.Println(err)
		}

	})
}

var c = oauth.NewConsumer(
	*consumerKey,
	*consumerSecret,
	oauth.ServiceProvider{
		RequestTokenUrl:   "https://api.twitter.com/oauth/request_token",
		AuthorizeTokenUrl: "https://api.twitter.com/oauth/authorize",
		AccessTokenUrl:    "https://api.twitter.com/oauth/access_token",
	})

func loading() {

	if len(*consumerKey) == 0 || len(*consumerSecret) == 0 {
		fmt.Println("You must set the --consumerkey and --consumersecret flags.")
		fmt.Println("---")
		os.Exit(1)
	}

	requestToken, url, err := c.GetRequestTokenAndUrl("oob")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("(1) Go to: " + url)
	fmt.Println("(2) Grant access, you should get back a verification code.")
	fmt.Println("(3) Enter that verification code here: ")
	verificationCode := ""
	fmt.Scanln(&verificationCode)
	fmt.Println("loading successed.")
	accessToken, err = c.AuthorizeToken(requestToken, verificationCode)
	if err != nil {
		log.Fatal(err)
	}

}

func get_timeline(procLine func(b []byte)) {

	response, err := c.Get(
		"https://userstream.twitter.com/1.1/user.json",
		map[string]string{},
		accessToken)
	if err != nil {
		log.Fatal(err)
	}

	defer response.Body.Close()

	scanner := bufio.NewScanner(response.Body)
	for scanner.Scan() {
		go procLine([]byte(scanner.Text()))
	}
	if err = scanner.Err(); err != nil {
		fmt.Println(err)
	}

	statuses := []status{}
	_ = json.NewDecoder(response.Body).Decode(&statuses)
	for _, s := range statuses {
		fmt.Printf("@%v: %v\n", s.User.ScreenName, s.Text)

	}
}

func UpdateStatus(text string, inReplyToStatusId uint64) error {
	response, err := c.Post("https://api.twitter.com/1.1/statuses/update.json",
		map[string]string{"status": text, "in_reply_to_status_id": fmt.Sprint(inReplyToStatusId)}, accessToken)

	if err != nil {
		return err
	}

	defer response.Body.Close()

	return nil
}

func updateName(name string) error {

	response, err := c.Post("https://api.twitter.com/1.1/account/update_profile.json",
		map[string]string{"name": name}, accessToken)
	if err != nil {
		log.Fatal(err)
	}
	defer response.Body.Close()

	return nil
}
