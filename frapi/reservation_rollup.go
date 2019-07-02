package frapi

import (
	"github.com/bjorge/friendlyreservations/models"
	"github.com/bjorge/friendlyreservations/utilities"
)

// ReservationRollup holds a snapshot of a reservation at each event
type ReservationRollup struct {
	// orginal reservation
	Input *models.NewReservationInput

	// rollup changes
	Canceled       bool
	UpdateDateTime string
	EventVersion   int32
}

// GetEventVersion returns version of rollup item
func (r *ReservationRollup) GetEventVersion() int {
	return int(r.EventVersion)
}

func (r *PropertyResolver) rollupReservations() {

	r.rollupMutexes[reservationRollupType].Lock()
	defer r.rollupMutexes[reservationRollupType].Unlock()

	if !r.rollupsExists(reservationRollupType) {

		for _, event := range r.property.Events {
			if newReservationInput, ok := event.(*models.NewReservationInput); ok {

				reservationRollup := &ReservationRollup{}
				reservationRollup.Input = newReservationInput
				reservationRollup.Canceled = false
				reservationRollup.UpdateDateTime = reservationRollup.Input.CreateDateTime

				reservationRollup.EventVersion = newReservationInput.EventVersion

				r.addRollup(newReservationInput.ReservationId,
					reservationRollup, reservationRollupType)
			}
			if cancelReservationInput, ok := event.(*models.CancelReservationInput); ok {
				ifaces := r.getRollups(&rollupArgs{id: &cancelReservationInput.ReservationId}, reservationRollupType)
				rollup, _ := ifaces[0].(*ReservationRollup)
				// make a copy of the rollup
				reservationRollup := *rollup

				// update the copy

				reservationRollup.Canceled = true
				reservationRollup.EventVersion = cancelReservationInput.EventVersion
				reservationRollup.UpdateDateTime = cancelReservationInput.CreateDateTime

				// store the copy as a new version of the rollup
				r.addRollup(cancelReservationInput.ReservationId,
					&reservationRollup, reservationRollupType)

			}
		}
		cacheError := r.cacheRollup(reservationRollupType)
		if cacheError != nil {
			utilities.LogWarningf(r.ctx, "cache write reservation rollups error: %+v", cacheError)
		}
	}

}
