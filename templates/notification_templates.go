package templates

// CurrentTemplateVersion is the current template version
const CurrentTemplateVersion = 1

// TemplateName is the type used for template names
type TemplateName string

// Template names
const (
	TestTemplate                  TemplateName = "TEST_TEMPLATE"
	NewPropertyNotification       TemplateName = "NOTIFICATION_NEW_PROPERTY"
	NewReservationNotification    TemplateName = "NEW_RESERVATION"
	CancelReservationNotification TemplateName = "CANCEL_RESERVATION"
	BalanceChangeNotification     TemplateName = "BALANCE_INCREASE"
	HomePageContents              TemplateName = "HOME_PAGE"
	LowBalanceNotification        TemplateName = "BALANCE_NOTIFICATION"
)

// TemplateParamGroup is the type used for template group names
type TemplateParamGroup string

// Template group names
const (
	Settings    TemplateParamGroup = "Settings"
	Reservation TemplateParamGroup = "Reservation"
	Ledger      TemplateParamGroup = "Ledger"
	Me          TemplateParamGroup = "Me"
	Decimal     TemplateParamGroup = "Decimal"
)

// GetNotificationTemplate returns two templates (ex. subject+body notification, or member+admin page)
func GetNotificationTemplate(version int, name TemplateName) (string, string, []TemplateParamGroup) {
	switch name {
	case LowBalanceNotification:
		return `{{.Settings.PropertyName}}: Negative balance reminder`,
			`Hi {{.User.Nickname}},
Just a reminder that you have a negative balance of {{.Ledger.Balance .Decimal}}.
		
Please submit a payment to cover your negative balance soon.
		
Thanks!
{{.Settings.PropertyName}}
`,
			[]TemplateParamGroup{Me, Ledger, Settings, Decimal}

	case TestTemplate:
		return "",
			`{{.Settings.PropertyName}}`,
			[]TemplateParamGroup{Settings}
	case HomePageContents:
		return `Welcome {{.Settings.PropertyName}} Member!
Hi {{.Me.Nickname}},
Your balance is {{.Ledger.Balance .Decimal}} today.

If your balance drops below {{.Settings.MinBalance .Decimal }} then you cannot make new reservations.

You can view balance details, make a reservation, etc. by selecting the menu button on the upper left.

Have Fun!
`,
			`Administrative Home

Select administrative options by pressing the menu button on the upper left.
`,
			[]TemplateParamGroup{Me, Ledger, Settings, Decimal}
	case NewPropertyNotification:
		return `{{.Settings.PropertyName}}: New property created`,
			`New property created`,
			[]TemplateParamGroup{Settings}
	case NewReservationNotification:
		return `{{.Settings.PropertyName}}: New reservation with check in on {{.Reservation.StartDate}} for {{.Reservation.ReservedFor.Nickname}}`,
			`Hi {{.Settings.PropertyName}} Members!

A new reservation has been made for check in on {{.Reservation.StartDate}} and check out on {{.Reservation.EndDate}}.

{{.Settings.PropertyName}}
`,
			[]TemplateParamGroup{Settings, Reservation}
	case CancelReservationNotification:
		return `{{.Settings.PropertyName}}: Reservation canceled with checkin date {{.Reservation.StartDate}}`,
			`Hi {{.Settings.PropertyName}} Members!
	
The reservation with checkin date {{.Reservation.StartDate}} has been canceled.
	
{{.Settings.PropertyName}}
`,
			[]TemplateParamGroup{Settings, Reservation}
	case BalanceChangeNotification:
		return `{{.Settings.PropertyName}}: Balance change for {{.User.Nickname}}`,
			`Hi {{.User.Nickname}},
	
Your balance has been changed by {{.Ledger.Amount .Decimal}} to a new balance of {{.Ledger.Balance .Decimal}}.

{{.Settings.PropertyName}}
`,
			[]TemplateParamGroup{Me, Ledger, Settings, Decimal}
	default:
		return "",
			"",
			nil
	}

}
