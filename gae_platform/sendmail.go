package gaeplatform

import (
	"context"

	"github.com/bjorge/friendlyreservations/platform"
	"google.golang.org/appengine/mail"
)

type sendMailImpl struct{}

// NewEmailSender is the factory method to create an email sender
func NewEmailSender() platform.SendMail {
	logging.LogDebugf("Email sender using gae implementation")
	return &sendMailImpl{}
}

func (r *sendMailImpl) Send(ctx context.Context, emailMessage *platform.EmailMessage) error {
	logging.LogDebugf("Send: sending email on gae platform")
	msg := &mail.Message{
		Sender:   emailMessage.Sender,
		ReplyTo:  emailMessage.ReplyTo,
		To:       emailMessage.To,
		Cc:       emailMessage.Cc,
		Bcc:      emailMessage.Bcc,
		Subject:  emailMessage.Subject,
		Body:     emailMessage.Body,
		HTMLBody: emailMessage.HTMLBody,
	}

	if emailMessage.Attachments != nil {
		attachments := []mail.Attachment{}
		for _, attachment := range emailMessage.Attachments {
			gaeAttachement := mail.Attachment{
				Name:      attachment.Name,
				Data:      attachment.Data,
				ContentID: attachment.ContentID,
			}
			attachments = append(attachments, gaeAttachement)
		}
		msg.Attachments = attachments
	}

	err := mail.Send(ctx, msg)
	if err == nil {
		logging.LogErrorf("Success sending email: %+v", msg.Subject)
	} else {
		logging.LogErrorf("Error sending mail: %+v", err)
	}
	return err
}
