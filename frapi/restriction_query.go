package frapi

import (
	"sort"

	"github.com/bjorge/friendlyreservations/frdate"

	"github.com/bjorge/friendlyreservations/models"
)

const restrictionGQL = `

# See BlackoutRestrictionInput.
type BlackoutRestriction {
	startDate: String!
	endDate: String!
}

type MembershipRestriction {
	prePayStartDate(format: String = "Jan 2, 2006"): String!
	inDate(format: String = "Jan 2, 2006"): String!
	outDate(format: String = "Jan 2, 2006"): String!
	gracePeriodOutDate(format: String = "Jan 2, 2006"): String!
	amount(format: AmountFormat = DECIMAL): String!
}

# See FirstDayRestrictionInput, LastDayRestrictionInput, BlackoutRestrictionInput.
union Restriction = BlackoutRestriction | MembershipRestriction

# See restriction inputs for descriptions.
type RestrictionRecord {
	restrictionId: String!
	createDateTime: String!
	author: User!
	description: String!
	restriction: Restriction!
}
`

type restrictionsArgs struct {
	RestrictionID *string
	MaxVersion    *int32
}

// Restrictions is called to return the list of restrictions
func (r *PropertyResolver) Restrictions(args *restrictionsArgs) ([]*RestrictionRecordResolver, error) {

	r.rollupRestrictions()

	// get rollups (with common filters applied)
	var l []*RestrictionRecordResolver
	settings, _ := r.Settings(&settingsArgs{MaxVersion: args.MaxVersion})
	dateBuilder := frdate.MustNewDateBuilder(settings.Timezone())
	ifaces := r.getRollups(&rollupArgs{id: args.RestrictionID, maxVersion: args.MaxVersion}, restrictionRollupType)
	for _, iface := range ifaces {
		resolver := &RestrictionRecordResolver{}
		resolver.property = r
		resolver.dateBuilder = dateBuilder
		resolver.args = args
		resolver.rollup = iface.(*RestrictionRollup)
		l = append(l, resolver)
	}

	sort.Slice(l, func(i, j int) bool {
		return l[i].GetEventVersion() < l[j].GetEventVersion()
	})

	return l, nil
}

// RestrictionResolver is the receiver for restriction conversion methods
type RestrictionResolver struct {
	restriction interface{}
	dateBuilder *frdate.DateBuilder
}

// ToBlackoutRestriction check and convert restriction to blackout restriction
func (r *RestrictionResolver) ToBlackoutRestriction() (*BlackoutRestrictionResolver, bool) {
	obj, ok := r.restriction.(*models.BlackoutRestriction)
	if ok {
		return &BlackoutRestrictionResolver{obj}, true
	}
	return nil, ok
}

// ToMembershipRestriction check and convert restriction to membership restriction
func (r *RestrictionResolver) ToMembershipRestriction() (*MembershipRestrictionResolver, bool) {
	obj, ok := r.restriction.(*models.MembershipRestriction)
	if ok {
		return &MembershipRestrictionResolver{obj, r.dateBuilder}, true
	}
	return nil, ok
}

// RestrictionRecordResolver holds underlying information for restriction resolvers
type RestrictionRecordResolver struct {
	rollup      *RestrictionRollup
	property    *PropertyResolver
	args        *restrictionsArgs
	dateBuilder *frdate.DateBuilder
}

// Restriction returns a restriction resolver for types of restrictions
func (r *RestrictionRecordResolver) Restriction() *RestrictionResolver {
	var iface interface{}
	if r.rollup.Input.Blackout != nil {
		iface = r.rollup.Input.Blackout
	} else {
		iface = r.rollup.Input.Membership
	}
	return &RestrictionResolver{iface, r.dateBuilder}
}

// RestrictionID is the unique restriction id
func (r *RestrictionRecordResolver) RestrictionID() string {
	return r.rollup.Input.RestrictionId
}

// CreateDateTime is the create time stamp of the restriction
func (r *RestrictionRecordResolver) CreateDateTime() string {
	return r.rollup.Input.CreateDateTime
}

// Author is the user that creates the restriction
func (r *RestrictionRecordResolver) Author() *UserResolver {
	users := r.property.Users(&usersArgs{UserID: &r.rollup.Input.AuthorUserId, MaxVersion: r.args.MaxVersion})
	return users[0]
}

// Description is the description of the restriction
func (r *RestrictionRecordResolver) Description() string {
	return r.rollup.Input.Description
}

// GetEventVersion is the version of this restriction record
func (r *RestrictionRecordResolver) GetEventVersion() int {
	return int(r.rollup.Input.EventVersion)
}

// BlackoutRestrictionResolver is a blackout resolver
type BlackoutRestrictionResolver struct {
	restriction *models.BlackoutRestriction
}

// EndDate is the last checkout date for a blackout restriction
func (r *BlackoutRestrictionResolver) EndDate() string {
	return r.restriction.EndDate
}

// StartDate is the first checkin date for a blackout restriction
func (r *BlackoutRestrictionResolver) StartDate() string {
	return r.restriction.StartDate
}

// MembershipRestrictionResolver is a membership resolver
type MembershipRestrictionResolver struct {
	restriction *models.MembershipRestriction
	dateBuilder *frdate.DateBuilder
}

// PrePayStartDate is the first date in which a membership can be purchased
func (r *MembershipRestrictionResolver) PrePayStartDate(args *struct{ Format string }) (string, error) {
	return formatDate(r.dateBuilder, r.restriction.PrePayStartDate, args.Format)
}

func (r *MembershipRestrictionResolver) internalPrePayStartDate() string {
	return r.restriction.PrePayStartDate
}

// InDate is the first allowed checkin date for a membership
func (r *MembershipRestrictionResolver) InDate(args *struct{ Format string }) (string, error) {
	return formatDate(r.dateBuilder, r.restriction.InDate, args.Format)
}

func (r *MembershipRestrictionResolver) inDateInternal() string {
	return r.restriction.InDate
}

// OutDate is the last allowed checkout date for a membership
func (r *MembershipRestrictionResolver) OutDate(args *struct{ Format string }) (string, error) {
	return formatDate(r.dateBuilder, r.restriction.OutDate, args.Format)
}

func (r *MembershipRestrictionResolver) internalOutDate() string {
	return r.restriction.OutDate
}

// GracePeriodOutDate is the last allowed checkout date for membership past the OutDate
func (r *MembershipRestrictionResolver) GracePeriodOutDate(args *struct{ Format string }) (string, error) {
	return formatDate(r.dateBuilder, r.restriction.GracePeriodOutDate, args.Format)
}

func (r *MembershipRestrictionResolver) internalGracePeriodOutDate() string {
	return r.restriction.GracePeriodOutDate
}

// Amount is the price of the membership
func (r *MembershipRestrictionResolver) Amount(args *struct{ Format amountFormat }) (string, error) {
	return formatAmount(r.restriction.Amount, args.Format)
}

func (r *MembershipRestrictionResolver) internalAmount() int32 {
	return r.restriction.Amount
}
