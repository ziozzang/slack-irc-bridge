# slack-to-irc bridge

This code is for slack to IRC bridge, also vice versa. You can build personal IRC client with slack. Original code is forked from https://github.com/voldyman/slack-irc-bridge.

## not implemented

not implemented IRC's user listing browsing. also, no DCC or chat function. only for generic channel chat.

## building

just clone source, ```go get``` and ```go build```.

## Usage

```shell
# using environment variables
./slack-irc-bridge

# with a custom config file
./slack-irc-bridge [configuration_file]
```

If you want to use slack as your personal IRC client, -> make channel in slack, mapping to irc channel. (set off relay option)
if you want to migrate from IRC to slack, -> make this code run as just bridge. also, you can build automatic invite script.

## Configuration

If no config file parameter is specified, configuration is pulled from environment variables.

### Environment Variables

The following environment variables are respected:

- IRC_URL (default: `irc://username:password@irc.freenode.org:6667?relay_nick=true`): A DSN containing connection information for irc.
- SLACK_URL (default: none, required: `true`): A slack workspace url in the form `https://example.slack.com`.
- SLACK_TOKEN (default: none, required: `true`): A slack bot token.
- BRIDGES (default: none, required: `true`): A comma separate list of colon separated slack/irc mappings in the form `#slack-channel:#irc-channel`.

### Config File

default file name is bot.config. if you want to use another name, just add file name as command parameter like ```./slack-irc-bridge foo.cfg```

bot.config file is json format. you can check it sample.

* You can get your slack token from https://api.slack.com/web

```
{
    "bridges":[
        {"slack":"#general", "irc":"#foo"},
        {"slack":"#slack-channel2", "irc": "#bar"}
    ],
    "irc": {
        "server": "irc.freenode.org:6667",
        "nick" : "bot-nickname",
        "relay_nick": true
    },
    "slack": {
        "url": "https://foo.slack.com",
        "token": "token_here"
    }
}
```
* irc/relay-nick is bool option. if you set true, bot will show your slack nick to IRC. if you use slack as single IRC client, turn off option.
* irc/server is IRC server option. this bot doesn't translate code page, so if you are in specific locale environment, use UTF-8 server.
