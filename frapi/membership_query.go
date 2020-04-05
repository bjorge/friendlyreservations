package frapi

import (
	"fmt"
)

type membershipsArgs struct {
	MaxVersion *int32
}

// Memberships exported for GQL
func (r *PropertyResolver) Memberships(args *membershipsArgs) ([]*MembershipResolver, error) {
	// rollup
	r.rollupMembershipStatus()

	versionedRollups := r.getRollups(&rollupArgs{maxVersion: args.MaxVersion}, membershipStatusRollupType)

	membershipResolvers := []*MembershipResolver{}

	users := r.Users(&usersArgs{MaxVersion: args.MaxVersion})
	usersMap := make(map[string]*UserResolver)
	for _, user := range users {
		usersMap[user.UserID()] = user
	}

	for _, rollup := range versionedRollups {

		membershipRollupRecord, ok := rollup.(*MembershipRollupRecord)
		if !ok {
			continue
		}

		membershipResolver := &MembershipResolver{}
		membershipResolver.membershipRollupRecord = membershipRollupRecord
		membershipResolver.users = usersMap

		membershipResolvers = append(membershipResolvers, membershipResolver)
	}

	return membershipResolvers, nil

}

// MembershipResolver exported for GQL
type MembershipResolver struct {
	membershipRollupRecord *MembershipRollupRecord
	users                  map[string]*UserResolver
}

// InDate is membership start date, exported for GQL
func (r *MembershipResolver) InDate() string {
	return r.membershipRollupRecord.InDate
}

// OutDate is membership end date, exported for GQL
func (r *MembershipResolver) OutDate() string {
	return r.membershipRollupRecord.OutDate
}

// Description is membership description, exported for GQL
func (r *MembershipResolver) Description() string {
	return r.membershipRollupRecord.Description
}

// Amount is membership price, exported for GQL
func (r *MembershipResolver) Amount() int32 {
	return r.membershipRollupRecord.Amount
}

// PrePayStartDate is membership pre pay start date, exported for GQL
func (r *MembershipResolver) PrePayStartDate() string {
	return r.membershipRollupRecord.PrePayStartDate
}

// GracePeriodOutDate is membership grace period out date, exported for GQL
func (r *MembershipResolver) GracePeriodOutDate() string {
	return r.membershipRollupRecord.GracePeriodOutDate
}

// MembershipStateResolver is user membership state, exported for GQL
type MembershipStateResolver struct {
	user            *UserResolver
	membershipState MembershipState
}

// MembershipStates are the user membership states for the membership, exported for GQL
func (r *MembershipResolver) MembershipStates() []*MembershipStateResolver {
	membershipStateResolvers := []*MembershipStateResolver{}
	for userID, state := range r.membershipRollupRecord.Users {
		membershipStateResolver := &MembershipStateResolver{}
		membershipStateResolver.user = r.users[userID]
		membershipStateResolver.membershipState = state
		membershipStateResolvers = append(membershipStateResolvers, membershipStateResolver)
	}
	for existingUserID, existingUser := range r.users {
		found := false
		if existingUser.IsSystem() {
			continue
		}
		for _, memberStateResolver := range membershipStateResolvers {
			if memberStateResolver.user.UserID() == existingUserID {
				found = true
				break
			}
		}
		if !found {
			membershipStateResolver := &MembershipStateResolver{}
			membershipStateResolver.user = existingUser
			membershipStateResolver.membershipState = OPEN
			membershipStateResolvers = append(membershipStateResolvers, membershipStateResolver)
		}

	}
	return membershipStateResolvers
}

// User is the user component of a membership state, exported for GQL
func (r *MembershipStateResolver) User() *UserResolver {
	return r.user
}

// State is the state component of a membership state, exported for GQL
func (r *MembershipStateResolver) State() string {
	return string(r.membershipState)
}

func (r *PropertyResolver) membershipInfo(restrictionID string, userID string, maxVersion int32) (MembershipState, error) {

	// rollup
	r.rollupMembershipStatus()

	versionedRollups := r.getRollups(&rollupArgs{maxVersion: &maxVersion, id: &restrictionID}, membershipStatusRollupType)

	if len(versionedRollups) == 0 {
		return OPEN, fmt.Errorf("restriction id is not valid")
	}

	rollup := versionedRollups[0].(*MembershipRollupRecord)

	if status, ok := rollup.Users[userID]; ok {
		return status, nil
	}

	return OPEN, nil
}
