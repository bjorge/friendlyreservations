package frapi

import (
	"bytes"
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"strconv"

	"github.com/bjorge/friendlyreservations/frdate"

	"github.com/bjorge/friendlyreservations/utilities"

	"github.com/bjorge/friendlyreservations/platform"
)

// ExportCSV is called to export CSV files for the current property
func (r *Resolver) ExportCSV(ctx context.Context, args *struct {
	PropertyID string
}) (*PropertyResolver, error) {
	// get the current property
	property, me, err := currentProperty(ctx, args.PropertyID)
	if err != nil {
		return nil, err
	}

	// check the input values
	if !me.IsAdmin() {
		return nil, errors.New("only admins can export")
	}

	// create the email message for the csv export
	msg, err := r.exportCSVInternal(ctx, property, me)
	if err != nil {
		return nil, err
	}

	// send the email
	err = EmailSender.Send(ctx, msg)

	return property, nil
}

func (r *Resolver) exportCSVInternal(ctx context.Context, property *PropertyResolver, me *UserResolver) (*platform.EmailMessage, error) {

	ledgersAttachment, err := r.exportLedgers(ctx, property, me)
	if err != nil {
		return nil, err
	}

	reservationsAttachment, err := r.exportReservations(ctx, property, me)
	if err != nil {
		return nil, err
	}

	sender := fmt.Sprintf("%s <%s>", utilities.SystemName, utilities.SystemEmail)
	to := []string{fmt.Sprintf("%s <%s>", me.Nickname(), me.Email())}

	msg := &platform.EmailMessage{
		Sender:      sender,
		To:          to,
		Subject:     "CSV export",
		Body:        "Attached are the CSV files",
		Attachments: []platform.EmailAttachment{*ledgersAttachment, *reservationsAttachment},
	}

	return msg, err
}

func (r *Resolver) exportLedgers(ctx context.Context, property *PropertyResolver, me *UserResolver) (*platform.EmailAttachment, error) {

	// get the setting to find the property timezone
	settings, _ := property.Settings(&settingsArgs{})
	db := frdate.MustNewDateBuilder(settings.Timezone())

	// get the ledgers
	ledgers, _ := property.Ledgers(&ledgersArgs{})

	var records [][]string

	records = append(records, []string{"member", "year", "date", "amount", "balance", "event"})
	for _, ledgerRecord := range ledgers {
		member := ledgerRecord.User()
		year, balance := 0, "0.00"
		for _, item := range ledgerRecord.Records() {
			if member.IsSystem() {
				continue
			}
			date := db.MustNewDateTime(item.EventDateTime()).ToDate()

			if year != 0 && year != date.Year() {
				record := []string{}
				record = append(record, member.Nickname())
				record = append(record, strconv.Itoa(year))
				record = append(record, strconv.Itoa(year)+"-12-31")
				record = append(record, "0.00")
				record = append(record, balance)
				record = append(record, "FINAL_BALANCE")
				records = append(records, record)
			}
			year = date.Year()
			balance = item.balanceInternal().Decimal()

			record := []string{}
			record = append(record, member.Nickname())
			record = append(record, strconv.Itoa(date.Year()))
			record = append(record, date.ToString())
			record = append(record, item.amountInternal().Decimal())
			record = append(record, item.balanceInternal().Decimal())
			record = append(record, string(eventName(item.Event())))
			records = append(records, record)
		}
		if !member.IsSystem() {
			record := []string{}
			record = append(record, member.Nickname())
			record = append(record, strconv.Itoa(year))
			record = append(record, strconv.Itoa(year)+"-12-31")
			record = append(record, "0.00")
			record = append(record, balance)
			record = append(record, "FINAL_BALANCE")
			records = append(records, record)
		}
	}

	// write the ledger attachment
	stream := &bytes.Buffer{}
	w := csv.NewWriter(stream)
	w.WriteAll(records) // calls Flush internally

	Logger.LogDebugf("CSV ledgers:\n%+v", string(stream.Bytes()))

	err := w.Error()
	if err != nil {
		return nil, err
	}

	attachment := platform.EmailAttachment{}
	//attachment.ContentID = utilities.NewGuid()
	attachment.Data = stream.Bytes()
	attachment.Name = "ledgers.csv"

	return &attachment, err

}

func (r *Resolver) exportReservations(ctx context.Context, property *PropertyResolver, me *UserResolver) (*platform.EmailAttachment, error) {

	// get the setting to find the property timezone
	settings, _ := property.Settings(&settingsArgs{})
	db := frdate.MustNewDateBuilder(settings.Timezone())

	// get the reservations
	reservations, _ := property.Reservations(&reservationsArgs{})

	var records [][]string

	records = append(records, []string{"member", "checkin year", "checkin date", "checkout date", "rate type", "daily rate", "days", "total amount", "author", "purchase year", "purchase date", "reservation id"})
	for _, reservationRecord := range reservations {

		if reservationRecord.rollup.Canceled {
			continue
		}

		// checkin and checkout dates
		checkinDate := db.MustNewDate(reservationRecord.StartDate())
		checkoutDate := db.MustNewDate(reservationRecord.EndDate())

		// check if reservation crosses a year boundary
		var numRecords int
		var checkinDateSplit *frdate.Date
		var checkoutDateSplit *frdate.Date
		_, firstDateNextYear := checkinDate.YearInOut()
		if !checkoutDate.After(firstDateNextYear) {
			numRecords = 1
		} else {
			numRecords = 2
			checkinDateSplit = firstDateNextYear
			checkoutDateSplit = checkoutDate
			checkoutDate = firstDateNextYear
			if checkoutDate.Year()-checkinDate.Year() > 1 {
				return nil, errors.New("Cannot export reservation that crosses more than 1 year")
			}
		}

		// get basic information about the reservation
		author := reservationRecord.Author().Nickname()
		reservedFor := reservationRecord.ReservedFor().Nickname()
		purchaseDate := db.MustNewDateTime(reservationRecord.UpdateDateTime())
		id := reservationRecord.ReservationID()

		// get the reservation rate at the time of making the reservation
		eventSettings, _ := property.Settings(&settingsArgs{MaxVersion: &reservationRecord.rollup.EventVersion})
		var rateType string
		var rate int32
		if reservationRecord.Member() {
			rateType = "MEMBER"
			rate = eventSettings.memberRateInternal()
		} else {
			rateType = "NONMEMBER"
			rate = eventSettings.nonMemberRateInternal()
		}

		record := []string{}
		record = append(record, reservedFor)
		record = append(record, strconv.Itoa(checkinDate.Year()))
		record = append(record, checkinDate.ToString())
		record = append(record, checkoutDate.ToString())
		record = append(record, rateType)
		record = append(record, currencyToString(int(rate)))
		record = append(record, strconv.Itoa(checkoutDate.Sub(checkinDate)))
		record = append(record, currencyToString(int(rate)*checkoutDate.Sub(checkinDate)))
		record = append(record, author)
		record = append(record, strconv.Itoa(purchaseDate.ToDate().Year()))
		record = append(record, purchaseDate.ToDate().ToString())
		record = append(record, id)
		records = append(records, record)

		if numRecords > 1 {
			record = []string{}
			record = append(record, reservedFor)
			record = append(record, strconv.Itoa(checkinDateSplit.Year()))
			record = append(record, checkinDateSplit.ToString())
			record = append(record, checkoutDateSplit.ToString())
			record = append(record, rateType)
			record = append(record, currencyToString(int(rate)))
			record = append(record, strconv.Itoa(checkoutDateSplit.Sub(checkinDateSplit)))
			record = append(record, currencyToString(int(rate)*checkoutDateSplit.Sub(checkinDateSplit)))
			record = append(record, author)
			record = append(record, strconv.Itoa(purchaseDate.ToDate().Year()))
			record = append(record, purchaseDate.ToDate().ToString())
			record = append(record, id)
			records = append(records, record)
		}
	}

	// write the ledger attachment
	stream := &bytes.Buffer{}
	w := csv.NewWriter(stream)
	w.WriteAll(records) // calls Flush internally

	Logger.LogDebugf("CSV reservations:\n%+v", string(stream.Bytes()))

	err := w.Error()
	if err != nil {
		return nil, err
	}

	attachment := platform.EmailAttachment{}
	//attachment.ContentID = utilities.NewGuid()
	attachment.Data = stream.Bytes()
	attachment.Name = "reservations.csv"

	return &attachment, err

}

func currencyToString(value int) string {
	p := strconv.Itoa(value)
	index := len(p) - 2
	q := p[:index] + "." + p[index:]
	return q
}

func eventName(event LedgerEvent) string {
	switch event {
	case paymentLedgerEvent:
		return "PAYMENT_CREDIT"
	case expenseLedgerEvent:
		return "EXPENSE_DEBIT"
	case reservationLedgerEvent:
		return "RESERVATION_PURCHASED"
	case cancelReservationLedgerEvent:
		return "RESERVATION_CANCELED"
	case purchaseMembershipLedgerEvent:
		return "MEMBERSHIP_PURCHASED"
	case optoutMembershipLedgerEvent:
		return "MEMBERSHIP_OPTOUT"
	case startLedgerEvent:
		return "NEW_MEMBER"
	default:
		panic("unknown event")
	}
}
