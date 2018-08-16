# SLACK WORKSPACE TERMINAL
## This is mostly meant for personal use and is subject to change/improvement. 

1. go get github.com/zkynet/slack-terminal (or git clone)
2. cd {GO SOUCE}/zkynet/slack-terminal/[windows/linux/mac]
3. touch .env
4. edit .env file to fit your needs.

# .env
```
SSH_KEY= [ Your private key file locations ]
SSH_PORT= [ ssh port ]
SSH_COMMAND_PATH= [ ssh binary location on your computer ]
SSH_COMMAND_ARGS= [ ssh flags (-c) ]
SSH_USER= [ your ssh username ]
SLACK_API_KEY= [ slack api key ]
```


# usage

```
@my-slack-bot TAG HOST COMMAND
```

```
@my-slack-bot TAG HOSTNAME ping -c 4 172.217.20.110
```

```
@my-slack-bot TAG HOSTNAME netstat -tulpn
```

# example
```
@my-slack-bot remote my.dev.domain.com netstat -tulpn
```

```
@my-slack-bot local netstat -tulpn
```