package frapi

import (
	"fmt"

	"github.com/bjorge/friendlyreservations/platform"

	"github.com/bjorge/friendlyreservations/models"
)

// LedgerRollup is the rollup record for ledgers, exported for memcache
type LedgerRollup struct {
	UserID string
	Event  LedgerEvent

	EventDateTime string
	Balance       int32
	Amount        int32

	VersionedEvent *platform.VersionedEvent

	EventVersion int32
}

// GetEventVersion of ledger rollup
func (r *LedgerRollup) GetEventVersion() int {
	return int(r.EventVersion)
}

func (r *PropertyResolver) rollupLedgers() {
	// rollup records if not rolled up already
	r.rollupMutexes[ledgerRollupType].Lock()
	defer r.rollupMutexes[ledgerRollupType].Unlock()

	if !r.rollupsExists(ledgerRollupType) {

		for _, event := range r.property.Events {

			// make a copy of the event for reference
			currentEvent := event

			switch ledgerEvent := event.(type) {

			case *models.NewUserInput:

				record := &LedgerRollup{}

				record.UserID = ledgerEvent.UserId
				record.Amount = 0
				record.Balance = 0
				record.EventDateTime = ledgerEvent.CreateDateTime
				record.EventVersion = ledgerEvent.EventVersion
				record.Event = startLedgerEvent
				record.VersionedEvent = &currentEvent

				r.addRollup(record.UserID, record, ledgerRollupType)

			case *models.NewReservationInput:

				rollups := r.getRollups(&rollupArgs{id: &ledgerEvent.ReservedForUserId}, ledgerRollupType)

				// make a copy
				record := *rollups[0].(*LedgerRollup)

				reservations, _ := r.Reservations(&reservationsArgs{ReservationID: &ledgerEvent.ReservationId, MaxVersion: &ledgerEvent.EventVersion})

				reservation := reservations[0]

				record.Amount = -1 * reservation.Amount()
				record.Balance -= reservation.Amount()
				record.EventDateTime = reservation.CreateDateTime()
				record.EventVersion = ledgerEvent.EventVersion
				record.Event = reservationLedgerEvent
				record.VersionedEvent = &currentEvent

				r.addRollup(record.UserID, &record, ledgerRollupType)

			case *models.CancelReservationInput:

				rollups := r.getRollups(&rollupArgs{id: &ledgerEvent.ReservedForUserId}, ledgerRollupType)

				// make a copy
				record := *rollups[0].(*LedgerRollup)

				reservations, _ := r.Reservations(&reservationsArgs{MaxVersion: &ledgerEvent.EventVersion,
					ReservationID: &ledgerEvent.ReservationId,
				})

				reservation := reservations[0]

				record.Amount = reservation.Amount()
				record.Balance += reservation.Amount()
				record.EventDateTime = reservation.UpdateDateTime()
				record.EventVersion = ledgerEvent.EventVersion
				record.Event = cancelReservationLedgerEvent
				record.VersionedEvent = &currentEvent

				r.addRollup(record.UserID, &record, ledgerRollupType)

			case *models.UpdateBalanceInput:

				rollups := r.getRollups(&rollupArgs{id: &ledgerEvent.UpdateForUserId}, ledgerRollupType)

				// make a copy
				record := *rollups[0].(*LedgerRollup)

				if ledgerEvent.Increase {
					record.Event = paymentLedgerEvent
					record.Amount = ledgerEvent.Amount
					record.Balance += ledgerEvent.Amount
				} else {
					record.Event = expenseLedgerEvent
					record.Amount = -1 * ledgerEvent.Amount
					record.Balance -= ledgerEvent.Amount
				}

				record.VersionedEvent = &currentEvent

				record.EventDateTime = ledgerEvent.CreateDateTime
				record.EventVersion = ledgerEvent.EventVersion

				r.addRollup(record.UserID, &record, ledgerRollupType)

			case *models.UpdateMembershipStatusInput:
				rollups := r.getRollups(&rollupArgs{id: &ledgerEvent.UpdateForUserId}, ledgerRollupType)

				// make a copy
				record := *rollups[0].(*LedgerRollup)

				// get the last status for this user and membership amount
				lastStatus, err := r.membershipInfo(ledgerEvent.RestrictionId, ledgerEvent.UpdateForUserId,
					ledgerEvent.EventVersion-1)
				if err != nil {
					panic(fmt.Errorf("restriction for id %v has no info", ledgerEvent.RestrictionId))
				}

				restrictions, err := r.Restrictions(&restrictionsArgs{RestrictionID: &ledgerEvent.RestrictionId})
				if err != nil || len(restrictions) != 1 {
					panic(fmt.Errorf("restriction for id %v does not exist", ledgerEvent.RestrictionId))
				}
				restriction, ok := restrictions[0].Restriction().ToMembershipRestriction()
				if !ok {
					panic(fmt.Errorf("restriction for id %v is not a membership restriction", ledgerEvent.RestrictionId))
				}
				membershipAmount := restriction.internalAmount()

				if ledgerEvent.Purchase {
					record.Event = purchaseMembershipLedgerEvent
					record.Amount = -1 * membershipAmount
					record.Balance -= membershipAmount
				} else {
					// ok, opt out, see if a purchase was made previously
					record.Event = optoutMembershipLedgerEvent
					if lastStatus == PURCHASED {
						record.Amount = membershipAmount
						record.Balance += membershipAmount
					} else {
						record.Amount = 0
					}

				}

				record.VersionedEvent = &currentEvent

				record.EventDateTime = ledgerEvent.CreateDateTime
				record.EventVersion = ledgerEvent.EventVersion

				r.addRollup(record.UserID, &record, ledgerRollupType)
			}
		}
		cacheError := r.cacheRollup(ledgerRollupType)
		if cacheError != nil {
			Logger.LogWarningf("cache write ledger rollups error: %+v", cacheError)
		}
	}
}
