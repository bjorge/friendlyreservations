package frapi

import (
	"bytes"
	"context"
	"encoding/gob"
	"errors"
	"fmt"

	"github.com/bjorge/friendlyreservations/utilities"
	"google.golang.org/appengine/mail"

	"github.com/bjorge/friendlyreservations/persist"
)

// PropertyExport defines the contents of an exported property gob
type PropertyExport struct {
	Events   []persist.VersionedEvent
	EmailMap map[string]string
}

// Export is called to export the current property
func (r *Resolver) Export(ctx context.Context, args *struct {
	PropertyID string
}) (*PropertyResolver, error) {
	// get the current property
	property, me, err := currentProperty(ctx, args.PropertyID)
	if err != nil {
		return nil, err
	}

	// check the input values
	if !me.IsAdmin() {
		return nil, errors.New("only admins can export")
	}

	msg, err := r.exportInternal(ctx, property, me)
	if err != nil {
		return nil, err
	}

	err = mail.Send(ctx, msg)
	if err != nil {
		utilities.LogErrorf(ctx, "Error sending mail: %+v", err)
	}

	return property, nil
}

func (r *Resolver) exportInternal(ctx context.Context, property *PropertyResolver, me *UserResolver) (*mail.Message, error) {

	// get the events
	export := &PropertyExport{}
	export.Events = property.property.Events
	emailMap := make(map[string]string)
	for _, record := range property.property.EmailMap {
		emailMap[record.EmailID] = record.Email
	}
	export.EmailMap = emailMap

	// encoded the array into a gob
	stream := &bytes.Buffer{}
	en := gob.NewEncoder(stream)
	err := en.Encode(export)
	if err != nil {
		return nil, err
	}

	sender := fmt.Sprintf("%s <%s>", utilities.SystemName, utilities.SystemEmail)
	to := []string{fmt.Sprintf("%s <%s>", me.Nickname(), me.Email())}

	attachment := mail.Attachment{}
	//attachment.ContentID = utilities.NewGuid()
	attachment.Data = stream.Bytes()
	attachment.Name = "frdatav1.bin"

	msg := &mail.Message{
		Sender:      sender,
		To:          to,
		Subject:     "Property export",
		Body:        "Attached is the export",
		Attachments: []mail.Attachment{attachment},
	}

	return msg, err
}
