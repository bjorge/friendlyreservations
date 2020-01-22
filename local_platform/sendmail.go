package localplatform

import (
	"context"

	"github.com/bjorge/friendlyreservations/platform"
)

type sendMailImpl struct{}

// NewEmailSender is the factory method to create an email sender
func NewEmailSender() platform.SendMail {

	return &sendMailImpl{}
}

func (r *sendMailImpl) Send(ctx context.Context, emailMessage *platform.EmailMessage) error {
	logging.LogErrorf("email sent to local platform: %+v", emailMessage)

	return nil
}
