package twilio_client

import (
	"context"
	"fmt"
	"os"

	"github.com/twilio/twilio-go"
	twilioapi "github.com/twilio/twilio-go/rest/api/v2010"
)

type TwilioClient interface {
	SendMessage(ctx context.Context, to string, body string, mediaUrl ...string) error
	SendMessageWithContactCard(ctx context.Context, to string) error
}

type Client struct {
	twilio *twilio.RestClient
}

func NewClient() (TwilioClient, error) {
	accountSid := os.Getenv("TWILIO_ACCOUNT_SID")
	authToken := os.Getenv("TWILIO_AUTH_TOKEN")

	twilioClient := twilio.NewRestClientWithParams(twilio.ClientParams{
		Username: accountSid,
		Password: authToken,
	})

	return &Client{twilio: twilioClient}, nil
}

func (c *Client) SendMessage(ctx context.Context, to string, body string, mediaUrl ...string) error {
	twilioPhoneNumber := os.Getenv("TWILIO_PHONE_NUMBER")

	params := twilioapi.CreateMessageParams{
		To:   &to,
		From: &twilioPhoneNumber,
		Body: &body,
	}

	if len(mediaUrl) > 0 {
		params.MediaUrl = &mediaUrl
	}

	_, err := c.twilio.Api.CreateMessage(&params)
	if err != nil {
		return fmt.Errorf("unable to send message: %w", err)
	}

	return nil
}

// TODOs:
// - add webhook to twilio to automatically send the contact card to new users
// upload vCard file
func (c *Client) SendMessageWithContactCard(ctx context.Context, to string) error {
	twilioPhoneNumber := os.Getenv("TWILIO_PHONE_NUMBER")
	contactCardUrl := os.Getenv("CONTACT_CARD_URL")

	body := "Save this contact for future Dodgers win notifications! 👇"

	params := twilioapi.CreateMessageParams{
		To:       &to,
		From:     &twilioPhoneNumber,
		Body:     &body,
		MediaUrl: &[]string{contactCardUrl},
	}

	_, err := c.twilio.Api.CreateMessage(&params)
	if err != nil {
		return fmt.Errorf("unable to send MMS with contact card: %w", err)
	}

	return nil
}
