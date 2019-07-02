package models

// import (
// 	"github.com/bjorge/friendlyreservations/persist"
// )

// rollups of time-series events returned to the client or used internally

// note: caching should be versioned to avoid issues...


type DailyRate struct {
	Date   string
	Amount int32
}

type UserState string

const (
	WAITING_ACCEPT UserState = "WAITING_ACCEPT"
	ACCEPTED       UserState = "ACCEPTED"
	DISABLED       UserState = "DISABLED"
	DECLINED       UserState = "DECLINED"
)
