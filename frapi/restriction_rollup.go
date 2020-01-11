package frapi

import (
	"github.com/bjorge/friendlyreservations/models"
)

// RestrictionRollup is an internal struct used duing rollup of restrictions
type RestrictionRollup struct {
	Input *models.NewRestrictionInput

	// rollup changes
	Canceled     bool
	EventVersion int32
}

// GetEventVersion returns version of rollup item
func (r *RestrictionRollup) GetEventVersion() int {
	return int(r.Input.EventVersion)
}

func (r *PropertyResolver) rollupRestrictions() {

	r.rollupMutexes[restrictionRollupType].Lock()
	defer r.rollupMutexes[restrictionRollupType].Unlock()

	// create a map of all the restriction events
	if !r.rollupsExists(restrictionRollupType) {

		for _, event := range r.property.Events {

			// for _, event := range events {
			newRestrictionInput, ok := event.(*models.NewRestrictionInput)

			if ok {

				restrictionRollup := &RestrictionRollup{}
				restrictionRollup.Input = newRestrictionInput
				restrictionRollup.EventVersion = newRestrictionInput.EventVersion
				restrictionRollup.Canceled = false

				r.addRollup(newRestrictionInput.RestrictionId,
					restrictionRollup, restrictionRollupType)
			}
		}
		cacheError := r.cacheRollup(restrictionRollupType)
		if cacheError != nil {
			Logger.LogWarningf("cache write restriction rollups error: %+v", cacheError)
		}
	}

}
