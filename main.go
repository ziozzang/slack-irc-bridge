package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"strings"

	// This does not include the NSB hack that was added in ziozzang's fork
	// This also adds WithLogin usage...
	"github.com/josegonzalez/ircbot"

	// Updates to latest nlopes/slack client
	slackbot "github.com/josegonzalez/slackbot-1"
)

type Bridges struct {
	Slack string `json:"slack"`
	IRC   string `json:"irc"`
}

type Slacks struct {
	Token string `json:"token"`
	URL   string `json:"url"`
}
type IRCs struct {
	Server    string `json:"server"`
	Nick      string `json:"nick"`
	Pass      string `json:"pass"`
	RelayNick bool   `json:"relay_nick"`
}
type Config struct {
	IRC    IRCs      `json:"irc"`
	Slack  Slacks    `json:"slack"`
	Bridge []Bridges `json:"bridges"`
}

func (c *Config) String() string {
	data, _ := json.Marshal(c)
	return string(data)
}

func getenv(key string, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		value = defaultValue
	}

	return value
}

func LoadConfigFromFile(s string) (*Config, error) {
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

func LoadConfigFromEnv() (*Config, error) {
	ircURL := getenv("IRC_URL", "irc://username:password@irc.freenode.org:6667?relay_nick=true")
	slackURL := getenv("SLACK_URL")
	slackToken := getenv("SLACK_TOKEN", "")
	bridgeConfigs := getenv("BRIDGES", "")

	if slackURL == "" {
		return nil, errors.New("SLACK_URL environment variable not specified")
	}
	if slackToken == "" {
		return nil, errors.New("SLACK_TOKEN environment variable not specified")
	}
	if bridgeConfigs == "" {
		return nil, errors.New("BRIDGES environment variable not specified")
	}

	var bridges []Bridges
	for _, bridgeConfig := range strings.Split(bridgeConfigs, ",") {
		channels := strings.SplitN(bridgeConfig, ":", 2)
		if len(channels) != 2 {
			return nil, errors.New(fmt.Sprintf("Invalid channel config for %s", bridgeConfig))
		}
		bridges = append(bridges, Bridges{channels[0], channels[1]})
	}

	slacks := Slacks{slackToken, slackURL}
	parsedURL, err := url.Parse(ircURL)
	if err != nil {
		return nil, err
	}

	username := parsedURL.User.Username()
	password, ok := parsedURL.User.Password()
	if !ok {
		password = ""
	}
	relayNick := parsedURL.Query().Get("relay_nick") == "true"
	server := fmt.Sprintf(parsedURL.Host)
	ircs := IRCs{server, username, password, relayNick}
	config := Config{ircs, slacks, bridges}
	return &config, nil
}

func getConfig() (*Config, error) {
	var filename string
	if len(os.Args[1:]) > 0 {
		filename = os.Args[1]
	} else {
		return LoadConfigFromEnv()
	}

	log.Println("loading configuration : ", filename)

	conf, err := LoadConfigFromFile(filename)
	if err != nil {
		log.Println("load configuration failed, err:", err)
		return nil, err
	}

	return conf, nil
}

func main() {
	conf, err := getConfig()
	if err != nil {
		log.Println("load configuration failed, err:", err)
		os.Exit(1)
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

	ircBot := ircbot.New(conf.IRC.Server, conf.IRC.Nick, conf.IRC.Pass, Values(bridges))
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
						ircBot.SendMessage(msg.Sender, msg.Text, target)
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
