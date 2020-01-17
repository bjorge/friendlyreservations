package persist

import (
	"os"
	"strconv"
	"time"

	"github.com/bjorge/friendlyreservations/logger"
)

var logging = logger.New()

// consolidateNumRecords means consolidate when this number of records exists in the datastore
var consolidateNumRecords int

// consolidateMaxSize means if a record is over this size then do not consolidate any more records to it
var consolidateMaxSize int

// consolidateCompress if true will compress events when storing
var consolidateCompress bool

var memcacheDuration time.Duration

func init() {
	value := os.Getenv("CONSOLIDATE_NUM_RECORDS")
	if value == "" {
		// default to 5
		consolidateNumRecords = 5
	} else {
		consolidateNumRecords, _ = strconv.Atoi(value)
	}

	value = os.Getenv("MEMCACHE_TIME_DURATION")
	if value == "" {
		// default to 5 min
		memcacheDuration, _ = time.ParseDuration("5m")
	} else {
		memcacheDuration, _ = time.ParseDuration(value)
	}

	value = os.Getenv("CONSOLIDATE_MAX_SIZE")
	if value == "" {
		// default to 90% of 1 MB
		consolidateMaxSize = 1048576 - 104857
	} else {
		consolidateMaxSize, _ = strconv.Atoi(value)
	}

	value = os.Getenv("CONSOLIDATE_COMPRESS")
	if value == "false" {
		consolidateCompress = false
	} else {
		// default to true
		consolidateCompress = true
	}

}
