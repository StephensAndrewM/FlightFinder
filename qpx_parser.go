package main

import (
    "time"
    "encoding/json"
    "net/http"
    "bytes"
    "fmt"
    "os"
    "strconv"
    "strings"
    "io/ioutil"
    "math"
    "github.com/mitchellh/hashstructure"
)

const QPX_URL = "https://www.googleapis.com/qpxExpress/v1/trips/search?key=" + API_KEY
const JSON_TYPE = "application/json"

type AppConfig struct {
    DryRun bool
    CacheOK bool
}

// QPX Request Items
type FlightsRequest struct {
    NumPassengers int
    Slices [2]FlightsRequestSlice
}

type FlightsRequestSlice struct {
    Origin string
    Destination string
    Date time.Time
    TimeBounds [2]string
    MaxLegs int
}

type FlightsResultOptionList []FlightsResultOption

type FlightsResult struct {
    Options FlightsResultOptionList
    Success bool
}

type FlightsResultOption struct {
    Price float64
    Slices [2]FlightsResultSlice
}

type FlightsResultSlice struct {
    Duration time.Duration
    Segments []FlightsResultSegment
}

type FlightsResultSegment struct {
    Airline string
    FlightNumber string
    Origin string
    Destination string
    DepartureTime time.Time
    ArrivalTime time.Time
    NumLegs int
}



type QPXRequest struct {
    Request QPXRequestContent      `json:"request"`
}
type QPXRequestContent struct {
    Passengers QPXPassengerCounts       `json:"passengers"`
    Slice [2]QPXSliceInput `json:"slice,omitempty"`
}
type QPXPassengerCounts struct {
    AdultCount int `json:"adultCount"`
}
type QPXSliceInput struct {
    Origin string `json:"origin"`
    Destination string `json:"destination"`
    Date string `json:"date"`
    MaxStops *int `json:"maxStops,omitempty"`
    PermittedDepartureTime *QPXTimeOfDayRange `json:"permittedDepartureTime,omitempty"`
}
type QPXTimeOfDayRange struct {
    EarliestTime string `json:"earliestTime"`
    LatestTime string `json:"latestTime"`
}


// QPX Response Items
type QPXResult struct {
    Trips QPXTripsResult    `json:"trips"`
    Error QPXResultError    `json:"error"`
}

type QPXTripsResult struct {
    Data       QPXData                  `json:"data"`
    TripOption []QPXTripOption                  `json:"tripOption"`
}

type QPXData struct {
    Airport []QPXAirport                    `json:"airport"`
    Carrier []QPXCarrier                    `json:"carrier"`
}

type QPXAirport struct {
    City string                 `json:"city"`
}

type QPXCarrier struct {
    Code string                 `json:"code"`
    Name string                 `json:"name"`
}

type QPXTripOption struct {
    SaleTotal string                    `json:"saleTotal"`
    Slice     []QPXSlice                    `json:"slice"`
}

type QPXSlice struct {
    Duration int                    `json:"duration"`
    Segment []QPXSegment                    `json:"segment"`
}

type QPXSegment struct {
    Flight QPXFlightDetail                  `json:"flight"`
    Leg    []QPXLeg                 `json:"leg"`
}

type QPXFlightDetail struct {
    Carrier string                  `json:"carrier"`
    Number  string                 `json:"number"`
}

type QPXLeg struct {
    ArrivalTime string                   `json:"arrivalTime"`
    DepartureTime string                 `json:"departureTime"`
    Origin string                   `json:"origin"`
    Destination string                  `json:"destination"`
}

type QPXResultError struct {
    Errors []QPXResultErrorEntry        `json:"errors"`
}

type QPXResultErrorEntry struct {
    Domain string       `json:"domain"`
    Reason string       `json:"reason"`
    Message string      `json:"message"`
}

// Helper Methods
func BuildQPXRequest(req FlightsRequest) (qpxReq QPXRequest) {

    const DATE_FMT = "2006-01-02"
    
    qpxReq.Request.Passengers.AdultCount = req.NumPassengers
    for i := 0; i < 2; i++ {
        qpxReq.Request.Slice[i].Origin = req.Slices[i].Origin
        qpxReq.Request.Slice[i].Destination = req.Slices[i].Destination
        qpxReq.Request.Slice[i].Date = req.Slices[i].Date.Format(DATE_FMT)

        if req.Slices[i].TimeBounds[0] != "" || req.Slices[i].TimeBounds[1] != "" {
            timeRange := new(QPXTimeOfDayRange)
            timeRange.EarliestTime = req.Slices[i].TimeBounds[0]
            timeRange.LatestTime = req.Slices[i].TimeBounds[1]
            qpxReq.Request.Slice[i].PermittedDepartureTime = timeRange
        }
        
        if req.Slices[i].MaxLegs > 0 {
            maxStops := new(int)
            *maxStops = req.Slices[i].MaxLegs-1
            qpxReq.Request.Slice[i].MaxStops = maxStops
        }
    }
    return qpxReq

}

func MakeQPXRequest(qpxReq QPXRequest, config AppConfig) (qpxRes QPXResult, success bool) {

    // fmt.Printf("QPX Request: %+v\n", qpxReq)

    // Encode the request struct as a JSON bytestring, then convert to buffer
    reqEncoded, encodingError := json.Marshal(qpxReq)
    if encodingError != nil {
        fmt.Printf("Error creating JSON for QPX request: %s\n", encodingError)
    }
    reqBuf := bytes.NewBuffer(reqEncoded)

    // fmt.Println(reqBuf.String())

    // Create a hash of the request object (for cache purposes)
    hash, hashError := hashstructure.Hash(qpxReq, nil)
    if hashError != nil {
        fmt.Printf("Error creating hash for QPX request: %s\n", hashError)
    }

    if config.DryRun {
        fmt.Println("Would have sent QPX Request: ")
        fmt.Printf("%+v\n", reqBuf.String())
        success = false
        return
    }

    cacheFile := "cache/qpx-"+strconv.FormatUint(hash, 10)

    // Results of getting flights JSON
    resBuf := new(bytes.Buffer)
    var jsonError error
    isCacheableResponse := false

    file, fileError := ioutil.ReadFile(cacheFile)
    if fileError != nil || !config.CacheOK {

        // fmt.Printf("Cache miss: %s\n", fileError)
        res, httpError := http.Post(QPX_URL, JSON_TYPE, reqBuf)
        if httpError != nil {
            fmt.Printf("Error communicating with QPX. Err: %s\n", httpError)
            success = false
            return
        }
        
        resBuf.ReadFrom(res.Body)
        jsonError = json.Unmarshal(resBuf.Bytes(), &qpxRes)
        isCacheableResponse = true

    } else {

        // fmt.Printf("Cache hit: %s\n", cacheFile)
        jsonError = json.Unmarshal(file, &qpxRes)

    }

    
    // fmt.Printf("QPX Response: %+v\n", resBuf)
    
    if jsonError != nil {
        fmt.Printf("Error interpreting QPX response. Err: %s\n", jsonError)
        success = false
        return
    }

    if len(qpxRes.Error.Errors) > 0 {
        success = false
        return
    }

    if isCacheableResponse {
        // Only write cache on success
        ioutil.WriteFile(cacheFile, resBuf.Bytes(), os.FileMode(770))
    }

    success = true
    return

}

func InterpretQPXResult(qpxRes QPXResult, success bool) (res FlightsResult) {

    var err error

    if !success {
        res.Success = false
        return
    }    

    res.Success = true

    const DATETIME_FMT = "2006-01-02T15:04-07:00"
    const DATETIMEOUT_FMT = "Mon Jan 02 03:04 PM MST"

    for _,qpxOption := range qpxRes.Trips.TripOption {
        var option FlightsResultOption
        option.Price = GetCurrencyValue(qpxOption.SaleTotal)

        var outboundSlice,inboundSlice FlightsResultSlice
        outboundSlice.Duration = time.Duration(qpxOption.Slice[0].Duration)*time.Minute
        inboundSlice.Duration = time.Duration(qpxOption.Slice[1].Duration)*time.Minute

        for _,qpxSegment := range qpxOption.Slice[0].Segment {
            var segment FlightsResultSegment
            segment.Airline = CarrierCodeToName(qpxSegment.Flight.Carrier, qpxRes.Trips.Data.Carrier)
            segment.FlightNumber = qpxSegment.Flight.Carrier + " " + qpxSegment.Flight.Number
            segment.Origin = qpxSegment.Leg[0].Origin
            segment.Destination = qpxSegment.Leg[len(qpxSegment.Leg) - 1].Destination
            segment.DepartureTime, err = time.Parse(DATETIME_FMT, qpxSegment.Leg[0].DepartureTime)
            if err != nil {
                fmt.Printf("Could not interpret departure date: %s\n", qpxSegment.Leg[0].DepartureTime)
            }
            segment.ArrivalTime, err = time.Parse(DATETIME_FMT, qpxSegment.Leg[len(qpxSegment.Leg) - 1].ArrivalTime)
            if err != nil {
                fmt.Printf("Could not interpret arrival date: %s", qpxSegment.Leg[len(qpxSegment.Leg) - 1].ArrivalTime)
            }
            segment.NumLegs = len(qpxSegment.Leg)
            outboundSlice.Segments = append(outboundSlice.Segments, segment)
        }

        for _,qpxSegment := range qpxOption.Slice[1].Segment {
            var segment FlightsResultSegment
            segment.Airline = CarrierCodeToName(qpxSegment.Flight.Carrier, qpxRes.Trips.Data.Carrier)
            segment.FlightNumber = qpxSegment.Flight.Carrier + " " + qpxSegment.Flight.Number
            segment.Origin = qpxSegment.Leg[0].Origin
            segment.Destination = qpxSegment.Leg[len(qpxSegment.Leg) - 1].Destination
            segment.DepartureTime, err = time.Parse(DATETIME_FMT, qpxSegment.Leg[0].DepartureTime)
            if err != nil {
                fmt.Printf("Could not interpret departure date: %s\n", qpxSegment.Leg[0].DepartureTime)
            }
            segment.ArrivalTime, err = time.Parse(DATETIME_FMT, qpxSegment.Leg[len(qpxSegment.Leg) - 1].ArrivalTime)
            if err != nil {
                fmt.Printf("Could not interpret arrival date: %s", qpxSegment.Leg[len(qpxSegment.Leg) - 1].ArrivalTime)
            }
            segment.NumLegs = len(qpxSegment.Leg)
            inboundSlice.Segments = append(inboundSlice.Segments, segment)
        }

        option.Slices[0] = outboundSlice
        option.Slices[1] = inboundSlice

        res.Options = append(res.Options, option)
    }
    return
}

func CarrierCodeToName(code string, lookup []QPXCarrier) (string) {
    for _,carrier := range lookup {
        if code == carrier.Code {
            return carrier.Name
        }
    }
    return "Unknown"
}

func GetDurationFromString(minutes string) (time.Duration) {
    i, err := strconv.Atoi(minutes)
    if err != nil {
        fmt.Println("Could not interpret duration: " + minutes)
        os.Exit(1)
    }
    return time.Duration(i)*time.Minute
}

// Expected Input Format: USD316.40
func GetCurrencyValue(amountStr string) (float64) {
    amountFl, err := strconv.ParseFloat(strings.Replace(amountStr, "USD", "", 1), 32)
    if err != nil {
        fmt.Println("Could not interpret float: " + amountStr)
        os.Exit(1)
    }
    return amountFl
}



func (o FlightsResultOption) getTripLength() (time.Duration) {
    tripStart := o.Slices[0].Segments[0].DepartureTime
    tripEnd := o.Slices[1].Segments[len(o.Slices[1].Segments)-1].ArrivalTime
    return tripEnd.Sub(tripStart)    
}

func (o FlightsResultOption) getPreferredAirlineScore() (score int) {
    for _,slice := range o.Slices {
        for _,segment := range slice.Segments {
            if strings.Contains(segment.Airline, "Jetblue") {
                score++
            }
        }
    }
    return
}

func (o FlightsResultOption) getPrice() (price int) {
    return int(math.Floor(o.Price/10)*10)
}


func (options FlightsResultOptionList) Len() int {
    return len(options)
}

func (options FlightsResultOptionList) Less(i, j int) bool {
    if options[i].getPrice() == options[j].getPrice() {
        if options[i].getTripLength() == options[j].getTripLength() {
            return options[i].getPreferredAirlineScore() > options[j].getPreferredAirlineScore()
        } else {
            return options[i].getTripLength() > options[j].getTripLength()
        }
    } else {
        return options[i].getPrice() < options[j].getPrice();        
    }
    
}

func (options FlightsResultOptionList) Swap(i, j int) {
    options[i], options[j] = options[j], options[i]
}