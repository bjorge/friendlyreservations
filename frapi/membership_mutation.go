package frapi

import (
	"context"
	"fmt"

	"github.com/bjorge/friendlyreservations/frdate"
	"github.com/bjorge/friendlyreservations/models"
	"github.com/bjorge/friendlyreservations/utilities"
)

// UpdateMembershipStatus is called to update membership status (ex. payed, opted out, etc.)
func (r *Resolver) UpdateMembershipStatus(ctx context.Context, args *struct {
	PropertyID string
	Input      *models.UpdateMembershipStatusInput
}) (*PropertyResolver, error) {
	utilities.DebugLog(ctx, "Update Membership Status")

	// get the current property
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

	// only allow admins if requested
	if args.Input.AdminUpdate && !me.IsAdmin() {
		return nil, fmt.Errorf("only an admin can make this call")
	}

	// comment required if admin request
	if args.Input.AdminUpdate && args.Input.Comment == nil {
		return nil, fmt.Errorf("comment required for admin update")
	}

	// validate comment string
	if args.Input.Comment != nil {
		args.Input.Comment, err = trim(*args.Input.Comment)
		if err != nil {
			return nil, err
		}
	}

	// get the restriction
	restrictions, err := property.Restrictions(&restrictionsArgs{
		RestrictionID: &args.Input.RestrictionId,
	})

	if len(restrictions) != 1 {
		return nil, fmt.Errorf("restriction not found for id: %+v", args.Input.RestrictionId)
	}

	if _, ok := restrictions[0].Restriction().ToMembershipRestriction(); !ok {
		return nil, fmt.Errorf("restriction is not for a membership: %+v", args.Input.RestrictionId)
	}

	// get the user
	users := property.Users(&usersArgs{UserID: &args.Input.UpdateForUserId})

	if len(users) != 1 {
		return nil, fmt.Errorf("user not found for id: %+v", args.Input.UpdateForUserId)
	}

	// check for dups

	// check that this state transition is valid

	// store the event!
	args.Input.AuthorUserId = me.UserID()
	args.Input.CreateDateTime = frdate.CreateDateTimeUTC()
	return commitChanges(ctx, args.PropertyID, property.EventVersion(), args.Input)
}
