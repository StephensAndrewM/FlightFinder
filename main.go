package main

import (
    "time"
    // "github.com/davecgh/go-spew/spew"
    "fmt"
    "sort"
    "github.com/fatih/color"
)

var DryRun bool
var CacheOK bool



func main() {

    input := InputParams{
        OriginAirports: []string{"SFO", "SJC"},
        DestAirport: "BOS",
        Outbound: DirectionParams{
            DateRange: [2]string{"2017-03-22", "2017-03-24"},
            RedEyeOnly: true,
        },
        Inbound: DirectionParams{
            DateRange: [2]string{"2017-03-26", "2017-03-27"},
        },
        MaxTripLength: 5,
        NumPassengers: 1,
        DryRun: false,
        CacheOK: true,
    }

    /*input := InputParams{
        OriginAirport: "BDL",
        DestAirports: []string{"SFO", "SJC"},
        Outbound: DirectionParams{
            DateRange: [2]string{"2017-03-29", "2017-03-31"},
        },
        Inbound: DirectionParams{
            DateRange: [2]string{"2017-04-02", "2017-04-03"},
        },
        NumPassengers: 2,
        MinTripLength: 4,
        DryRun: false,
        CacheOK: true,
    }*/

    DryRun = input.DryRun
    CacheOK = input.CacheOK

    // spew.Dump(input.GetOriginAirports())
    // spew.Dump(input.GetDestAirports())
    // spew.Dump(input.GetValidDateRanges())
    
    reqList := BuildFlightRequest(input)
    // spew.Dump(reqList)
    resList := makeParallelQPXRequests(reqList)
    options, successes := FlattenResponses(resList)

    successFont := color.New(color.FgGreen, color.Bold)
    failureFont := color.New(color.FgRed, color.Bold)

    if successes == len(reqList) {
        successFont.Printf("All %d queries return successfully!\n", successes)
    } else {
        failureFont.Printf("Errors! Only %d/%d queries returned successfully.\n", successes, len(reqList))
    }

    
    PrintResults(options)

}

func BuildFlightRequest(input InputParams) (reqList []FlightsRequest) {

    for _,outboundOrigin := range input.GetOriginAirports() {
        for _,outboundDest := range input.GetDestAirports() {
            for _,inboundOrigin := range input.GetDestAirports() {
                for _,inboundDest := range input.GetOriginAirports() {
                    for _,dateRange := range input.GetValidDateRanges() {
                        var req FlightsRequest
                        req.NumPassengers = input.NumPassengers
                        req.Slices[0] = FlightsRequestSlice{
                            Origin: outboundOrigin,
                            Destination: outboundDest,
                            Date: dateRange[0],
                            TimeBounds: input.Outbound.GetTimeRange(),
                            MaxLegs: input.Outbound.GetMaxLegs(),
                        }
                        req.Slices[1] = FlightsRequestSlice{
                            Origin: inboundOrigin,
                            Destination: inboundDest,
                            Date: dateRange[1],
                            TimeBounds: input.Inbound.GetTimeRange(),
                            MaxLegs: input.Inbound.GetMaxLegs(),
                        }
                        // spew.Dump(req)
                        reqList = append(reqList, req)
                    }
                }
            }
        }
    }
    return
}

func makeParallelQPXRequests(reqList []FlightsRequest) (resList []FlightsResult) {

    c := make(chan FlightsResult, len(reqList))
    processed := 0
    limiter := time.Tick(time.Millisecond * 200)

    for _,req := range reqList {
        <-limiter  // Don't overload QPX
        go parallelQPXRequestHandler(req, c)
    }

    for i := 0; i < len(reqList); i++ {
        res := <-c
        resList = append(resList, res)
        processed++
        fmt.Printf("Received %d Out of %d Responses\n", processed, len(reqList))
    }

    return

}

func parallelQPXRequestHandler(req FlightsRequest, c chan FlightsResult) {

    qpxReq := BuildQPXRequest(req)
    qpxRes, success := MakeQPXRequest(qpxReq)
    res := InterpretQPXResult(qpxRes, success)
    c <- res

}

func FlattenResponses(resList []FlightsResult) (optionsList FlightsResultOptionList, successes int) {

    for _,result := range resList {
        if result.Success {
            successes++
            optionsList = append(optionsList, result.Options...)
        }
    }
    sort.Sort(optionsList)
    return

}

