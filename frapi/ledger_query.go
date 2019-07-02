package frapi

import (
	"fmt"
	"sort"
	"strconv"

	"github.com/bjorge/friendlyreservations/models"
)

const ledgerQueryGQL = `

enum LedgerEvent {
	PAYMENT
	EXPENSE
	RESERVATION
	CANCEL_RESERVATION
	MEMBERSHIP_PAYMENT
	MEMBERSHIP_OPTOUT
	START
}

type LedgerRecord {
	eventDateTime: String!
	event: LedgerEvent!
	balance(format: AmountFormat = DECIMAL): String!
	amount(format: AmountFormat = DECIMAL): String!
}

type Ledger {
	user: User!
	records: [LedgerRecord]!
}

`

// LedgerEvent is the ledger record event, exported for GQL
type LedgerEvent string

const (
	paymentLedgerEvent            LedgerEvent = "PAYMENT"
	expenseLedgerEvent            LedgerEvent = "EXPENSE"
	reservationLedgerEvent        LedgerEvent = "RESERVATION"
	cancelReservationLedgerEvent  LedgerEvent = "CANCEL_RESERVATION"
	purchaseMembershipLedgerEvent LedgerEvent = "MEMBERSHIP_PAYMENT"
	optoutMembershipLedgerEvent   LedgerEvent = "MEMBERSHIP_OPTOUT"
	startLedgerEvent              LedgerEvent = "START"
)

type ledgersArgs struct {
	UserID     *string
	Last       *int32
	Reverse    *bool
	MaxVersion *int32
}

// Ledgers is called by gql to query user ledgers
func (r *PropertyResolver) Ledgers(args *ledgersArgs) ([]*LedgerResolver, error) {

	// validate
	if err := r.validateLedgersInput(args); err != nil {
		return nil, err
	}

	// rollup
	r.rollupLedgers()

	// resolve
	return r.resolveLedgers(args)
}

// LedgerResolver exported for GQL
type LedgerResolver struct {
	userID   string
	rollups  []*LedgerRollup
	property *PropertyResolver
}

// User returns ledger user
func (r *LedgerResolver) User() *UserResolver {
	users := r.property.Users(&usersArgs{UserID: &r.userID})
	return users[0]
}

// Records returns ledger records for a user
func (r *LedgerResolver) Records() []*LedgerRecordResolver {
	l := []*LedgerRecordResolver{}
	for _, rollup := range r.rollups {
		recordResolver := &LedgerRecordResolver{}
		recordResolver.rollup = rollup
		l = append(l, recordResolver)
	}
	return l
}

// LedgerRecordResolver resolves a ledger record
type LedgerRecordResolver struct {
	rollup *LedgerRollup
}

// Amount is the ledger record amount
func (r *LedgerRecordResolver) Amount(args *struct{ Format amountFormat }) (string, error) {
	return formatAmount(r.rollup.Amount, args.Format)
}

func (r *LedgerRecordResolver) amountInternal() *amountResolver {
	return &amountResolver{r.rollup.Amount}
}

// Event is the ledger record event
func (r *LedgerRecordResolver) Event() LedgerEvent {
	return r.rollup.Event
}

// UserID is the ledger record user ID
func (r *LedgerRecordResolver) UserID() string {
	return r.rollup.UserID
}

type amountResolver struct {
	amount int32
}

func (r *amountResolver) Raw() int32 {
	return r.amount
}

func (r *amountResolver) NoDecimal() string {
	return strconv.Itoa(int(r.amount))
}

func (r *amountResolver) Decimal() string {
	decimalPart := r.amount % 100
	nonDecimalPart := r.amount / 100
	return fmt.Sprintf("%d.%02d", int(nonDecimalPart), int(decimalPart))
}

// Balance is the ledger record balance
func (r *LedgerRecordResolver) Balance(args *struct{ Format amountFormat }) (string, error) {
	return formatAmount(r.rollup.Balance, args.Format)
}

func (r *LedgerRecordResolver) balanceInternal() *amountResolver {
	return &amountResolver{r.rollup.Balance}
}

// EventDateTime is the ledger record timestamp
func (r *LedgerRecordResolver) EventDateTime() string {
	return r.rollup.EventDateTime
}

// GetEventVersion returns the event version for the ledger record
func (r *LedgerRecordResolver) GetEventVersion() int {
	return int(r.rollup.EventVersion)
}

func (r *PropertyResolver) validateLedgersInput(args *ledgersArgs) error {

	me, _ := r.Me()

	if args.UserID == nil {
		if !me.IsAdmin() && !me.IsSystem() {
			return fmt.Errorf("must be admin or system (ex. cron job) to view all records")
		}
	} else {
		users := r.Users(&usersArgs{UserID: args.UserID})

		if len(users) != 1 {
			return fmt.Errorf("user with id %v does not exist", args.UserID)
		}

		if !me.IsAdmin() && users[0].UserID() != *args.UserID {
			return fmt.Errorf("user must be admin or same user to view ledger")
		}
	}

	if args.Last != nil && *args.Last <= 0 {
		return fmt.Errorf("last (%+v) must be > 0", *args.Last)
	}

	return nil
}

func (r *PropertyResolver) resolveLedgers(args *ledgersArgs) ([]*LedgerResolver, error) {

	allVersions := true
	ifaces := r.getRollups(&rollupArgs{id: args.UserID, maxVersion: args.MaxVersion, allVersions: &allVersions}, ledgerRollupType)

	// map each user to their list of ledgers
	mapUserToRollupList := make(map[string][]*LedgerRollup)
	for _, iface := range ifaces {
		rollup := iface.(*LedgerRollup)
		mapUserToRollupList[rollup.UserID] = append(mapUserToRollupList[rollup.UserID], rollup)
	}

	// create the resolvers
	resolvers := []*LedgerResolver{}
	for userID, rollups := range mapUserToRollupList {
		resolver := &LedgerResolver{}
		resolver.userID = userID
		resolver.rollups = rollups
		resolver.property = r

		reverseOrder := false
		if args.Reverse != nil {
			reverseOrder = *args.Reverse
		}

		sort.Slice(resolver.rollups, func(i, j int) bool {
			if reverseOrder {
				return resolver.rollups[i].GetEventVersion() > resolver.rollups[j].GetEventVersion()
			}
			return resolver.rollups[i].GetEventVersion() < resolver.rollups[j].GetEventVersion()

		})

		if args.Last != nil {
			last := int(*args.Last)
			if last < len(resolver.rollups) {
				// example len = 3, last = 1, return 2:3
				resolver.rollups = resolver.rollups[len(resolver.rollups)-last : len(resolver.rollups)]
			}
		}

		resolvers = append(resolvers, resolver)
	}

	return resolvers, nil
}

// lastActiveLedgerBalances returns map of userIds and balances (called by cron)
func (r *PropertyResolver) lastActiveLedgerBalances() map[string]int32 {
	one := int32(1)
	reverse := false
	ledgers, err := r.Ledgers(&ledgersArgs{Last: &one, Reverse: &reverse})
	if err != nil {
		return nil
	}
	mapLedgers := make(map[string]int32)
	for _, ledger := range ledgers {
		if !ledger.User().IsSystem() && ledger.User().State() == models.ACCEPTED {
			record := ledger.Records()[0]
			mapLedgers[record.UserID()] = record.balanceInternal().Raw()
		}
	}
	return mapLedgers
}
