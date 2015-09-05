package main

import (
	"fmt"
	"log"
	"os"
	"io/ioutil"
	"encoding/json"

	"./ircbot"
	"github.com/voldyman/slackbot"
)

type Bridges struct {
	Slack		string `json:"slack"`
	IRC		string `json:"irc"`
}

type Slacks struct {
	Token		string `json:"token"`
	URL		string `json:"url"`
}
type IRCs struct {
	Server		string `json:"server"`
	Nick		string `json:"nick"`
	RelayNick	bool `json:"relay_nick"`
}
type Config struct {
	IRC		IRCs `json:"irc"`
	Slack		Slacks `json:"slack"`
	Bridge		[]Bridges `json:"bridges"`
}

func (c *Config) String() string {
	data, _ := json.Marshal(c)
	return string(data)
}

func LoadConfig(s string) (*Config, error) {
	data, err := ioutil.ReadFile(s)
	if err != nil {
		return nil, err
	}
	cConfig := &Config{}
	if err = json.Unmarshal(data, cConfig); err != nil {
		return nil, err
	}
	return cConfig, nil
}

func main() {
	filename := "bot.config"
	if len(os.Args[1:]) > 0 {
		filename = os.Args[1]
	}

	log.Println("loading configuration : ", filename)

	conf, err := LoadConfig(filename)
	if err != nil {
		log.Println("load configuration failed, err:", err)
		return
        }

	bridges := map[string]string{}
	for _, m := range conf.Bridge {
		bridges[m.Slack] = m.IRC
	}

	users := make(map[string]int)

	slackBot := slackbot.New(conf.Slack.Token)

	slackEvents, err := slackBot.Start(conf.Slack.URL)
	if err != nil {
		fmt.Println("Could not start slack bot", err.Error())
		return
	}

	ircBot := ircbot.New(conf.IRC.Server, conf.IRC.Nick, Values(bridges))
	ircEvents, err := ircBot.Start()
	if err != nil {
		fmt.Println("Could not connect to IRC")
		return
	}

	for {
		select {
		case msg := <-ircEvents:
			log.Printf("IRC: <%s@%s> %s\n", msg.Sender, msg.Channel, msg.Text)

			if target, ok := KeyForValue(bridges, msg.Channel); ok {
				slackBot.SendMessage(msg.Sender, target, msg.Text)
				incUser(users, msg.Sender, target)
			}

		case ev := <-slackEvents:
			switch ev.(type) {

			case *slackbot.MessageEvent:
				msg := ev.(*slackbot.MessageEvent)

				// we don't handle named channels without '#'
				msg.Channel = "#" + msg.Channel

				if _, ok := bridges[msg.Channel]; !ok {
					continue
				}
				log.Printf("slack: <%s@%s> %s\n", msg.Sender,
					msg.Channel, msg.Text)

				if shouldHandle(users, msg.Sender, msg.Channel) {
					log.Println("Handling Message")

					if target, ok := bridges[msg.Channel]; ok {
						ircBot.SendMessage(msg.Sender, msg.Text, target, conf.IRC.RelayNick)
					}

				}

				//case error:
				//	err = ev.(error)
				//	fmt.Println("Error occured:", err.Error())

			}
		}
	}

}

// Get All the keys of the map
func Keys(bridges map[string]string) []string {
	vals := []string{}

	for k := range bridges {
		vals = append(vals, k)
	}

	return vals
}

// Get all values of the map
func Values(bridges map[string]string) []string {
	vals := []string{}

	for _, v := range bridges {
		vals = append(vals, v)
	}

	return vals
}

// Get the key of a map for the given value
func KeyForValue(bridges map[string]string, val string) (string, bool) {
	result := ""

	for k, v := range bridges {
		if v == val {
			result = k
		}
	}

	if result == "" {
		return "", false
	}
	return result, true
}

// Semaphores to manages messages

func incUser(users map[string]int, user, channel string) {
	key := user + channel
	if val, ok := users[key]; ok {
		users[key] = val + 1
	} else {
		users[key] = 1
	}
}

func shouldHandle(users map[string]int, user, channel string) bool {
	key := user + channel
	if val, ok := users[key]; ok {
		if val > 0 {
			users[key] = val - 1
			return false
		}
	}

	return true
}
