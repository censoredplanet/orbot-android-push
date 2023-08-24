package fcmsender

import (
	"context"
	"log"
	"sync"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"
)

type FCMSender struct {
	ctx       context.Context
	fcmServer *messaging.Client
}

// SendTo sends data to user identified by token in a goroutine, and calls
// wg.Done() upon finish
func (c *FCMSender) SendTo(data map[string]string, token string, wg *sync.WaitGroup) {

	// Ye Shu: do we need TTL=0 for PushRSS? These messages should be latency-friendly and do not become out-of-date frequently?
	//headers := map[string]string{}
	//headers["ttl"] = fmt.Sprint(0)

	message := &messaging.Message{
		//Webpush: &messaging.WebpushConfig{
		//	Headers: headers,
		//},
		Data:  data,
		Token: token,
	}

	go func() {
		_, err := c.fcmServer.Send(c.ctx, message)
		if err != nil {
			if messaging.IsQuotaExceeded(err) {
				log.Printf("fcm send quota exceeded: %v", err)
				// TODO: handle this. How?
			} else if messaging.IsUnregistered(err) {
				log.Printf("fcm send token invalid: %v", err)
				// TODO: delete this token
			} else {
				log.Printf("fcm send error unhandled: %v", err)
			}
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
