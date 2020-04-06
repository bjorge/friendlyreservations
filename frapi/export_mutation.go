package frapi

import (
	"bytes"
	"context"
	"encoding/gob"
	"errors"
	"fmt"

	"github.com/bjorge/friendlyreservations/utilities"

	"github.com/bjorge/friendlyreservations/platform"
)

// PropertyExport defines the contents of an exported property gob
type PropertyExport struct {
	Events   []platform.VersionedEvent
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

	// check if export backup is allowed
	constraints, err := property.UpdateSettingsConstraints(ctx)
	if err != nil {
		return nil, err
	}

	if !constraints.AllowPropertyExportBackup() {
		return nil, errors.New("settings constraints do not allow export backup")
	}

	// check the input values
	if !me.IsAdmin() {
		return nil, errors.New("only admins can export")
	}

	msg, err := r.exportInternal(ctx, property, me)
	if err != nil {
		return nil, err
	}

	err = EmailSender.Send(ctx, msg)

	return property, nil
}

func (r *Resolver) exportInternal(ctx context.Context, property *PropertyResolver, me *UserResolver) (*platform.EmailMessage, error) {

	// get the events
	export := &PropertyExport{}
	export.Events = property.property.Events
	emailMap := make(map[string]string)
	for key, value := range property.property.EmailMap {
		emailMap[key] = value
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

	attachment := platform.EmailAttachment{}
	//attachment.ContentID = utilities.NewGuid()
	attachment.Data = stream.Bytes()
	attachment.Name = "frdatav1.bin"

	msg := &platform.EmailMessage{
		Sender:      sender,
		To:          to,
		Subject:     "Property export",
		Body:        "Attached is the export",
		Attachments: []platform.EmailAttachment{attachment},
	}

	return msg, err
}
