package frapi

import (
	"context"

	"github.com/bjorge/friendlyreservations/frdate"
	"github.com/bjorge/friendlyreservations/models"
	"github.com/bjorge/friendlyreservations/platform"
	"github.com/bjorge/friendlyreservations/templates"
	"github.com/bjorge/friendlyreservations/utilities"
)

type notificationTargetType string

const (
	notificationTargetAdmins     notificationTargetType = "NOTIFICATION_TARGET_ADMINS"
	notificationTargetMember     notificationTargetType = "NOTIFICATION_TARGET_MEMBER"
	notificationTargetAllMembers notificationTargetType = "NOTIFICATION_TARGET_ALL_MEMBERS"
)

func sendEmail(ctx context.Context, property *PropertyResolver, notification *NotificationResolver) error {

	if !notification.rollup.Input.EmailSent {
		Logger.LogWarningf("sendEmail: notification email not configured to be sent")
		return nil
	}

	subject, err := notification.Subject()
	if err != nil {
		Logger.LogErrorf("sendEmail: Error resolving notification subject: %+v", err)
		return err
	}
	body, err := notification.Body()
	if err != nil {
		Logger.LogErrorf("sendEmail: Error resolving notification body: %+v", err)
		return err
	}

	msg := &platform.EmailMessage{
		Sender:  *notification.authorEmailFormat(),
		To:      notification.toEmailFormat(),
		Cc:      notification.ccEmailFormat(),
		Subject: subject,
		Body:    body,
	}

	return EmailSender.Send(ctx, msg)
}

func createNotificationRecord(notificationTarget notificationTargetType, property *PropertyResolver, templateName templates.TemplateName,
	singleTargetUserID *string, paramGroup *templates.TemplateParamGroup, paramData *string) *models.NewNotificationInput {

	users := property.Users(&usersArgs{})
	newNotification := &models.NewNotificationInput{}

	if utilities.SendMailDisabled {
		newNotification.EmailSent = false
	} else {
		newNotification.EmailSent = true
	}

	// setup the to and cc lists
	newNotification.CcUserIds = []string{}
	newNotification.ToUserIds = []string{}
	newNotification.AllNotifiedUserIds = []string{}
	for _, user := range users {
		if user.IsSystem() {
			newNotification.AuthorUserId = user.UserID()
		}

		switch notificationTarget {
		case notificationTargetAdmins:
			if user.IsAdmin() {
				if user.State() == models.ACCEPTED {
					// notify and send email
					newNotification.ToUserIds = append(newNotification.ToUserIds, user.UserID())
				} else if user.State() == models.WAITING_ACCEPT {
					// notify only
					newNotification.AllNotifiedUserIds = append(newNotification.AllNotifiedUserIds, user.UserID())
				}
			}

		case notificationTargetMember:
			if user.UserID() == *singleTargetUserID {
				// only user in the To list
				if user.State() == models.ACCEPTED {
					// notify and send email
					newNotification.ToUserIds = append(newNotification.ToUserIds, user.UserID())
				} else if user.State() == models.WAITING_ACCEPT {
					// notify only
					newNotification.AllNotifiedUserIds = append(newNotification.AllNotifiedUserIds, user.UserID())
				}
			} else if user.IsAdmin() {
				// all admins into the Cc list
				if user.State() == models.ACCEPTED {
					// notify and send email
					newNotification.CcUserIds = append(newNotification.CcUserIds, user.UserID())
				} else if user.State() == models.WAITING_ACCEPT {
					// notify only
					newNotification.AllNotifiedUserIds = append(newNotification.AllNotifiedUserIds, user.UserID())
				}
			}

		case notificationTargetAllMembers:
			if user.IsAdmin() && !user.IsMember() {
				// admin and not a member, so into the Cc list
				if user.State() == models.ACCEPTED {
					// notify and send email
					newNotification.CcUserIds = append(newNotification.CcUserIds, user.UserID())
				} else if user.State() == models.WAITING_ACCEPT {
					// notify only
					newNotification.AllNotifiedUserIds = append(newNotification.AllNotifiedUserIds, user.UserID())
				}
			} else if user.IsMember() {
				// all members and admin members into To list
				if user.State() == models.ACCEPTED {
					// notify and send email
					newNotification.ToUserIds = append(newNotification.ToUserIds, user.UserID())
				} else if user.State() == models.WAITING_ACCEPT {
					// notify only
					newNotification.AllNotifiedUserIds = append(newNotification.AllNotifiedUserIds, user.UserID())
				}
			}
		}
	}

	if len(newNotification.ToUserIds) == 0 {
		newNotification.ToUserIds = newNotification.CcUserIds
		newNotification.CcUserIds = []string{}
	}

	// add other To and Cc users into the all list
	for _, userID := range newNotification.ToUserIds {
		newNotification.AllNotifiedUserIds = append(newNotification.AllNotifiedUserIds, userID)
	}
	for _, userID := range newNotification.CcUserIds {
		newNotification.AllNotifiedUserIds = append(newNotification.AllNotifiedUserIds, userID)
	}

	newNotification.TemplateParamData = make(map[templates.TemplateParamGroup]string)
	if paramGroup != nil && paramData != nil {
		newNotification.TemplateParamData[*paramGroup] = *paramData
	}
	newNotification.TemplateName = templateName
	newNotification.CreateDateTime = frdate.CreateDateTimeUTC()
	newNotification.NotificationId = utilities.NewGUID()
	newNotification.TemplateVersion = int32(templates.CurrentTemplateVersion)
	newNotification.DefaultTemplate = true

	return newNotification
}
