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
	Dates []string
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

	// Create a list of the possible dates, based on what's given
	if len(input.Outbound.Date) > 0 {
		possibleOutboundDates = []time.Time{
			DateStringToTime(input.Outbound.Date) }
	} else if (len(input.Outbound.Dates) > 0) {
		possibleOutboundDates = DateListToTimeList(input.Outbound.Dates)
	} else {
		possibleOutboundDates = DateRangeToTimeList(
			input.Outbound.DateRange[0], input.Outbound.DateRange[1],
			input.Outbound.WeekdayExclusions)
	}

	if len(input.Inbound.Date) > 0 {
		possibleInboundDates = []time.Time{
			DateStringToTime(input.Inbound.Date) }
	} else if (len(input.Inbound.Dates) > 0) {
		possibleInboundDates = DateListToTimeList(input.Inbound.Dates)
	} else {
		possibleInboundDates = DateRangeToTimeList(
			input.Inbound.DateRange[0], input.Inbound.DateRange[1],
			input.Inbound.WeekdayExclusions)
	}

	// Calculate the min and max duration in seconds (how Go wants it)
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

	// Now interleave all the dates into pairs
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

func DateListToTimeList(dates []string) (validDates []time.Time) {
	for _, d := range dates {
		validDates = append(validDates, DateStringToTime(d))
	}
	return
}

func DateRangeToTimeList(start, end, dayRestrictions string) (validDates []time.Time) {

	dStart := DateStringToTime(start)
	dEnd := DateStringToTime(end)

	dEndUpperBound := dEnd.AddDate(0,0,1) // Add 1 day to make range inclusive
	for d := dStart; d.Before(dEndUpperBound); d = d.AddDate(0,0,1) {
		if IsDayAllowed(d.Weekday(), dayRestrictions) {
			validDates = append(validDates, d)
		}
	}

	return

}

func DateStringToTime(dateStr string) (time.Time) {
	const DATE_FMT = "2006-01-02"
	d, err := time.Parse(DATE_FMT, dateStr)
	if err != nil {
		fmt.Printf("Could not interpret date: %s\n", dateStr)
		os.Exit(1)
	}
	return d
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