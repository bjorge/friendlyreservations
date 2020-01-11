package frapi

import (
	"context"
	"fmt"

	"github.com/bjorge/friendlyreservations/frdate"
	"github.com/bjorge/friendlyreservations/models"
)

// CreateContent is the gql call to create new display content
func (r *Resolver) CreateContent(ctx context.Context, args *struct {
	PropertyID string
	Input      *models.NewContentInput
}) (*PropertyResolver, error) {
	Logger.LogDebugf("CreateContent")

	property, me, err := currentProperty(ctx, args.PropertyID)
	if err != nil {
		return nil, err
	}

	// check the input values and for duplicates
	if duplicate, err := isDuplicate(ctx, args.Input, property); duplicate || err != nil {
		if err == nil {
			return property, nil
		}
		return nil, err
	}

	if !me.IsAdmin() {
		return nil, fmt.Errorf("admin required for this call")
	}

	if args.Input == nil {
		return nil, fmt.Errorf("no args")
	}

	stringArg, err := trim(args.Input.Template)
	if err != nil {
		return nil, err
	}
	args.Input.Template = *stringArg

	stringArg, err = trim(args.Input.Comment)
	if err != nil {
		return nil, err
	}
	args.Input.Comment = *stringArg

	// input looks good, now add extra internal values
	args.Input.CreateDateTime = frdate.CreateDateTimeUTC()
	args.Input.AuthorUserId = me.UserID()

	// persist the event
	return commitChanges(ctx, args.PropertyID, property.EventVersion(), args.Input)
	//return commitChangesDeprecated(ctx, args.PropertyID, args.Input)
}
