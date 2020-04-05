package frapi

import (
	"github.com/bjorge/friendlyreservations/models"
)

// MembershipState is member's state for a particular membership
type MembershipState string

const (
	// OPEN means the membership is open for purchase or opt out
	// only one membership can be open at a time
	OPEN MembershipState = "OPEN"
	// PURCHASED means the member has purchased a membership
	// only a purchased membership can have reservations for a member
	PURCHASED MembershipState = "PURCHASED"
	// OPTOUT means the member has opted out, allowing another (future) membership to be open
	// opted out memberships cannot have reservations for a member
	OPTOUT MembershipState = "OPTOUT"
)

// MembershipRollupRecord is the rollup record for membership, exported for memcache
type MembershipRollupRecord struct {
	RestrictionID      string
	InDate             string
	OutDate            string
	PrePayStartDate    string
	GracePeriodOutDate string
	Users              map[string]MembershipState
	EventVersion       int32
	Description        string
	Amount             int32
}

// GetEventVersion returns the version of the rollup record
func (r *MembershipRollupRecord) GetEventVersion() int {
	return int(r.EventVersion)
}

func (r *PropertyResolver) rollupMembershipStatus() {

	// rollup records if not rolled up already
	r.rollupMutexes[membershipStatusRollupType].Lock()
	defer r.rollupMutexes[membershipStatusRollupType].Unlock()

	if !r.rollupsExists(membershipStatusRollupType) {

		for _, event := range r.property.Events {

			switch rollupEvent := event.(type) {

			case *models.NewRestrictionInput:

				if rollupEvent.Membership != nil {
					rollup := &MembershipRollupRecord{
						RestrictionID:      rollupEvent.RestrictionId,
						Users:              make(map[string]MembershipState),
						EventVersion:       rollupEvent.EventVersion,
						InDate:             rollupEvent.Membership.InDate,
						OutDate:            rollupEvent.Membership.OutDate,
						PrePayStartDate:    rollupEvent.Membership.PrePayStartDate,
						GracePeriodOutDate: rollupEvent.Membership.GracePeriodOutDate,
						Description:        rollupEvent.Description,
						Amount:             rollupEvent.Membership.Amount,
					}

					r.addRollup(rollupEvent.RestrictionId, rollup, membershipStatusRollupType)
				}

			case *models.UpdateMembershipStatusInput:
				rollups := r.getRollups(&rollupArgs{id: &rollupEvent.RestrictionId}, membershipStatusRollupType)

				if len(rollups) == 1 {
					last := rollups[0].(*MembershipRollupRecord)
					next := &MembershipRollupRecord{}
					next.EventVersion = rollupEvent.EventVersion
					next.RestrictionID = last.RestrictionID
					next.InDate = last.InDate
					next.OutDate = last.OutDate
					next.Description = last.Description
					next.Amount = last.Amount

					next.PrePayStartDate = last.PrePayStartDate
					next.GracePeriodOutDate = last.GracePeriodOutDate
					next.Users = make(map[string]MembershipState)
					for user, status := range last.Users {
						next.Users[user] = status
					}
					if rollupEvent.Purchase {
						next.Users[rollupEvent.UpdateForUserId] = PURCHASED
					} else {
						next.Users[rollupEvent.UpdateForUserId] = OPTOUT
					}
					r.addRollup(next.RestrictionID, next, membershipStatusRollupType)
				}
			}
		}
		cacheError := r.cacheRollup(membershipStatusRollupType)
		if cacheError != nil {
			Logger.LogWarningf("cache write membership rollups error: %+v", cacheError)
		}
	}
}
