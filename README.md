# Test for packet loss and send email alerts using Go


## Requirements

- Set ENV variables

```

Examples:
export SMTP_HOST="smtp.gmail.com"
export SMTP_PORT="587"
export SMTP_PASSWORD="asdfdsaf"
export SENDER_EMAIL="asdfdsf@gmail.com"
export DEST_EMAIL="asdfdasf@gmail.com"
export PACKET_LOSS_PERCENTAGE=20
```

Compiling for Raspberry pi

```
env GOOS=linux GOARCH=arm GOARM=6 go build
```
