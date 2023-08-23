package fcmsender

import (
	"context"
	"firebase.google.com/go"
	"firebase.google.com/go/messaging"
	"fmt"
	"google.golang.org/api/option"
	"log"
	"sync"
)

type FCMSender struct {
	ctx       context.Context
	fcmServer *messaging.Client
}

// SendTo sends data to user identified by token in a goroutine, and calls
// wg.Done() upon finish
func (c *FCMSender) SendTo(data map[string]string, token string, wg *sync.WaitGroup) {

	// TODO: Ye Shu: do we need TTL=0 for PushRSS? These messages should be latency-friendly and do not become out-of-date frequently?
	headers := map[string]string{}
	headers["ttl"] = fmt.Sprint(0)

	message := &messaging.Message{
		Webpush: &messaging.WebpushConfig{
			Headers: headers,
		},
		Data:  data,
		Token: token,
	}

	go func() {
		_, err := c.fcmServer.Send(c.ctx, message)
		if err != nil {
			log.Fatalf("fcm send error: %v", err)
		}
		wg.Done()
	}()
}

func NewFCMSender(credentialFilename string) *FCMSender {

	ctx := context.Background()
	opt := option.WithCredentialsFile(credentialFilename)
	config := firebase.Config{}
	app, err := firebase.NewApp(ctx, &config, opt)
	if err != nil {
		log.Fatalf("error new application: %v", err)
	}
	fcmserverapp, err := app.Messaging(ctx)
	if err != nil {
		log.Fatalf("error getting Messaging client: %v", err)
	}

	fcmSender := &FCMSender{
		ctx:       ctx,
		fcmServer: fcmserverapp,
	}
	return fcmSender
}
