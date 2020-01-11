package frapi

import (
	"context"
	"errors"

	"github.com/bjorge/friendlyreservations/frdate"
	"github.com/bjorge/friendlyreservations/models"
	"github.com/bjorge/friendlyreservations/utilities"
)

// CreateRestriction is called to create a new restriction
func (r *Resolver) CreateRestriction(ctx context.Context, args *struct {
	PropertyID string
	Input      *models.NewRestrictionInput
}) (*PropertyResolver, error) {
	Logger.LogDebugf("Create Restriction")

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

	if args.Input == nil {
		return nil, errors.New("missing restriction input arg")
	}

	if !me.IsAdmin() {
		return nil, errors.New("only an admin can create a restriction")
	}

	if args.Input.Description == "" {
		return nil, errors.New("the description is empty")
	}

	restrictions, err := property.Restrictions(&restrictionsArgs{})
	if err != nil {
		return nil, err
	}

	for _, restriction := range restrictions {
		if restriction.Description() == args.Input.Description {
			return nil, errors.New("the description matches an existing restriction")
		}
	}

	if args.Input.Blackout != nil && args.Input.Membership != nil {
		return nil, errors.New("only one restriction type is allowed")
	}

	settings, _ := property.Settings(&settingsArgs{})
	b, err := frdate.NewDateBuilder(settings.Timezone())
	if err != nil {
		return nil, err
	}

	if args.Input.Blackout != nil {
		startDate, err := b.NewDate(args.Input.Blackout.StartDate)
		if err != nil {
			return nil, errors.New("invalid blackout start date")
		}
		endDate, err := b.NewDate(args.Input.Blackout.EndDate)
		if err != nil {
			return nil, errors.New("invalid blackout end date")
		}
		if startDate.After(endDate) || startDate.ToString() == endDate.ToString() {
			return nil, errors.New("end date must be after start date")
		}
	}
	if args.Input.Membership != nil {
		inDate, err := b.NewDate(args.Input.Membership.InDate)
		if err != nil {
			return nil, errors.New("invalid membership in date")
		}
		outDate, err := b.NewDate(args.Input.Membership.OutDate)
		if err != nil {
			return nil, errors.New("invalid membership out date")
		}
		graceDate, err := b.NewDate(args.Input.Membership.GracePeriodOutDate)
		if err != nil {
			return nil, errors.New("invalid membership grace period date")
		}

		prePayStartDate, err := b.NewDate(args.Input.Membership.PrePayStartDate)
		if err != nil {
			return nil, errors.New("invalid membership pre pay start date")
		}

		if inDate.ToString() == outDate.ToString() || inDate.After(outDate) {
			return nil, errors.New("end date must be after start date")
		}

		if graceDate.Before(outDate) {
			return nil, errors.New("grace date cannot be before checkout date")
		}

		if prePayStartDate.After(inDate) {
			return nil, errors.New("pre pay date cannot be after first checkin date")
		}

		for _, restriction := range restrictions {
			if membership, ok := restriction.Restriction().ToMembershipRestriction(); ok {
				existingInDate, _ := b.NewDate(membership.inDateInternal())
				existingOutDate, _ := b.NewDate(membership.internalOutDate())
				if frdate.DateOverlapInOut(inDate, outDate, existingInDate, existingOutDate) {
					return nil, errors.New("membership overlaps existing membership")
				}
			}
		}
	}

	// update the request with more information
	args.Input.CreateDateTime = frdate.CreateDateTimeUTC()
	args.Input.RestrictionId = utilities.NewGUID()
	args.Input.AuthorUserId = me.UserID()

	// persist the event
	return commitChanges(ctx, args.PropertyID, property.EventVersion(), args.Input)
}
