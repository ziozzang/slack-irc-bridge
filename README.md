Slack-to-IRC Bridge
===================

This code is for slack to IRC bridge, also vice versa. You can build personal IRC client with slack.

Not Implemented
===============

a

Build
=====
just clone source, ```go get`` and ```go build```

after build, you must edit bot.config file to complete setup.

Configuration
=============

bot.config file is json format.

You can get your slack token from https://api.slack.com/web

Usage
=====

If you want to use slack as your personal IRC client, -> make channel in slack, mapping to irc channel. (set off relay option)
if you want to migrate from IRC to slack, -> make this code run as just bridge. also, you can build automatic invite script.

