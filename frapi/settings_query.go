package frapi

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/bjorge/friendlyreservations/frdate"
)

const settingsGQL = `
type Settings {
	propertyName: String!
	currency(format: CurrencyFormat = ACRONYM): String!
	memberRate(format: AmountFormat = DECIMAL): String!
	allowNonMembers: Boolean!
	nonMemberRate(format: AmountFormat = DECIMAL): String!
	timezone: String!
	minBalance(format: AmountFormat = DECIMAL): String!
	maxOutDays: Int!
	minInDays: Int!
	reservationReminderDaysBefore: Int!
	balanceReminderIntervalDays: Int!
}

enum AmountFormat {
	DECIMAL
	NODECIMAL
}

enum Currency {
	USD
	EUR
}

enum CurrencyFormat {
	ACRONYM
	SYMBOL
}

`

// there are not multiple settings, so just set a fixed id
const settingsID = "0"

type settingsArgs struct {
	MaxVersion *int32
}

// Settings method for GQL call
func (r *PropertyResolver) Settings(args *settingsArgs) (*SettingsResolver, error) {

	r.rollupSettings()

	id := settingsID
	ifaces := r.getRollups(&rollupArgs{id: &id, maxVersion: args.MaxVersion}, settingsRollupType)

	if len(ifaces) != 1 {
		return nil, errors.New("wrong number of settings records")
	}
	settings := ifaces[0].(*SettingsRollup)
	resolver := &SettingsResolver{settings: settings, property: r}
	return resolver, nil
}

// SettingsResolver is the settings resolver receiver
type SettingsResolver struct {
	settings *SettingsRollup
	property *PropertyResolver
}

// PropertyName is the name of the requested property
func (r *SettingsResolver) PropertyName() string {
	return r.settings.PropertyName
}

type currencyFormat string

const (
	acronym currencyFormat = "ACRONYM"
	symbol  currencyFormat = "SYMBOL"
)

// Currency method for GQL call
func (r *SettingsResolver) Currency(ctx context.Context, args *struct{ Format currencyFormat }) (string, error) {

	Logger.LogDebugf("currency format: %+v", args.Format)

	if args.Format == "" {
		args.Format = acronym
	}
	if args.Format == acronym {
		return string(r.settings.Currency), nil
	}
	if args.Format == symbol {
		switch r.settings.Currency {
		case "USD":
			return "$", nil
		case "EUR":
			return "€", nil
		}
	}

	return "", fmt.Errorf("unknown currency format")
}

type amountFormat string

const (
	decimal   amountFormat = "DECIMAL"
	nodecimal amountFormat = "NODECIMAL"
)

func formatDate(dateBuilder *frdate.DateBuilder, date string, format string) (string, error) {
	dateObj := dateBuilder.MustNewDate(date)
	formattedDate := dateObj.Format(format)
	return formattedDate, nil
}

func formatAmount(amount int32, format amountFormat) (string, error) {

	if format == "" {
		format = decimal
	}

	switch format {
	case decimal:
		decimalPart := amount % 100
		nonDecimalPart := amount / 100
		return fmt.Sprintf("%d.%02d", int(nonDecimalPart), int(decimalPart)), nil
	case nodecimal:
		return strconv.Itoa(int(amount)), nil
	}
	return "", fmt.Errorf("unknown format %+v", format)
}

// MemberRate is the daily member rate for reservations
func (r *SettingsResolver) MemberRate(args *struct{ Format amountFormat }) (string, error) {
	return formatAmount(r.settings.MemberRate, args.Format)
}

func (r *SettingsResolver) memberRateInternal() int32 {
	return r.settings.MemberRate
}

// AllowNonMembers is true if friends of members can have reservations
func (r *SettingsResolver) AllowNonMembers() bool {
	return r.settings.AllowNonMembers
}

// NonMemberRate is the daily non-member rate for reservations
func (r *SettingsResolver) NonMemberRate(args *struct{ Format amountFormat }) (string, error) {
	return formatAmount(r.settings.NonMemberRate, args.Format)
}

func (r *SettingsResolver) nonMemberRateInternal() int32 {
	return r.settings.NonMemberRate
}

// Timezone is the time zone of the property
func (r *SettingsResolver) Timezone() string {
	return r.settings.Timezone
}

// GetEventVersion returns the version of the settings event
func (r *SettingsResolver) GetEventVersion() int {
	return int(r.settings.EventVersion)
}

// MinBalance is the minimum balance required to make new reservations
func (r *SettingsResolver) MinBalance(args *struct{ Format amountFormat }) (string, error) {
	return formatAmount(r.settings.MinBalance, args.Format)
}

func (r *SettingsResolver) minBalanceInternal() *amountResolver {
	return &amountResolver{r.settings.MinBalance}
}

// MinInDays is the number of days before the current date when new reservations can be made
func (r *SettingsResolver) MinInDays() int32 {
	return r.settings.MinInDays
}

// MaxOutDays is the number of days after today after which new reservations are not allowed
func (r *SettingsResolver) MaxOutDays() int32 {
	return r.settings.MaxOutDays
}

// ReservationReminderDaysBefore is the number of days before the start of a reservation at which a reminder is sent
func (r *SettingsResolver) ReservationReminderDaysBefore() int32 {
	return r.settings.ReservationReminderDaysBefore
}

// BalanceReminderIntervalDays is the minimum interval between reminders for a negative balance
func (r *SettingsResolver) BalanceReminderIntervalDays() int32 {
	return r.settings.BalanceReminderIntervalDays
}
