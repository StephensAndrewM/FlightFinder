package main

import(
	"time"
	"fmt"
	"os"
	"math"
	"strings"
	// "github.com/davecgh/go-spew/spew"
)

type InputParams struct {
	OriginAirport    string
	OriginAirports []string
	DestAirport      string
	DestAirports   []string

	Outbound DirectionParams	
	Inbound DirectionParams

	NumPassengers int
	MinTripLength int
	MaxTripLength int
	DryRun bool
	CacheOK bool
}

type DirectionParams struct {
	Date string
	DateRange [2]string
	WeekdayExclusions string

	RedEyeOnly bool
	MaxLegs int
	TimeRange [2]string
}

func (input InputParams) GetOriginAirports() ([]string) {
	if len(input.OriginAirport) > 0 {
        return []string{ input.OriginAirport }
    } else {
    	return input.OriginAirports
    }
}

func (input InputParams) GetDestAirports() ([]string) {
	if len(input.DestAirport) > 0 {
        return []string{input.DestAirport}
    } else {
    	return input.DestAirports
    }
}

func (input InputParams) GetValidDateRanges() ([][]time.Time) {

	var possibleOutboundDates, possibleInboundDates []time.Time

	// Turn the upper and lower bounds of the dates into lists
	if len(input.Outbound.Date) > 0 {
		possibleOutboundDates = DateRangeToList(
			input.Outbound.Date, input.Outbound.Date, "")
	} else {
		possibleOutboundDates = DateRangeToList(
			input.Outbound.DateRange[0], input.Outbound.DateRange[1], input.Outbound.WeekdayExclusions)
	}
	if len(input.Inbound.Date) > 0 {
		possibleInboundDates = DateRangeToList(
			input.Inbound.Date, input.Inbound.Date, "")
	} else {
		possibleInboundDates = DateRangeToList(
			input.Inbound.DateRange[0], input.Inbound.DateRange[1], input.Inbound.WeekdayExclusions)
	}

	// Now interleave all the dates into pairs
	var minDuration, maxDuration time.Duration

    if input.MinTripLength > 0 {
        minDuration = time.Hour*time.Duration(24*input.MinTripLength)
    } else {
        minDuration = time.Duration(0)
    }
    if input.MaxTripLength > 0 {
        maxDuration = time.Hour*time.Duration(24*input.MaxTripLength)
    } else {
        maxDuration = time.Duration(math.MaxInt64)
    }

    var ranges [][]time.Time
    for _,outboundDate := range possibleOutboundDates {
        for _,inboundDate := range possibleInboundDates {
            if inboundDate.Sub(outboundDate) >= minDuration &&
                inboundDate.Sub(outboundDate) <= maxDuration {
                ranges = append(ranges, []time.Time{ outboundDate,inboundDate })
            }
        }
    }

    return ranges

}

func DateRangeToList(start, end, dayRestrictions string) (validDates []time.Time) {

	const DATE_FMT = "2006-01-02"

	var dStart, dEnd time.Time
	var err error
	dStart, err = time.Parse(DATE_FMT, start)
	if err != nil {
		fmt.Printf("Could not interpret date: %s\n", start)
		os.Exit(1)
	}
	dEnd, err = time.Parse(DATE_FMT, end)
	if err != nil {
		fmt.Printf("Could not interpret date: %s\n", end)
		os.Exit(1)
	}

	dEndUpperBound := dEnd.AddDate(0,0,1)
	for d := dStart; d.Before(dEndUpperBound); d = d.AddDate(0,0,1) {
		if IsDayAllowed(d.Weekday(), dayRestrictions) {
			validDates = append(validDates, d)
		}
	}

	return

}

func IsDayAllowed(d time.Weekday, dayRestrictions string) (bool) {
	if len(dayRestrictions) == 0 {
		return true
	}
	m := make(map[string]string)
	m["Sunday"] = "U"
	m["Monday"] = "M"
	m["Tuesday"] = "T"
	m["Wednesday"] = "W"
	m["Thursday"] = "R"
	m["Friday"] = "F"
	m["Saturday"] = "S"
	if weekdayCode, ok := m[d.String()]; ok {
		return !strings.Contains(dayRestrictions, weekdayCode)
	} else {
		return true
	}
}

func (direction DirectionParams) GetTimeRange() ([2]string) {

	if direction.RedEyeOnly {
		return [2]string{ "19:00", "23:59" }
	} else {
		return direction.TimeRange
	}

}

func (direction DirectionParams) GetMaxLegs() int {

	if direction.RedEyeOnly {
		return 1
	} else {
		return direction.MaxLegs
	}

}