package main

import (
    "time"
    // "github.com/davecgh/go-spew/spew"
    "fmt"
    "sort"
)


func main() {

    // Input is defined in input.go

    config := AppConfig{
        DryRun: Input.DryRun,
        CacheOK: Input.CacheOK,
    }

    reqList := BuildFlightRequest(Input)
    resList := MakeParallelQPXRequests(reqList, config)
    options, successes := FlattenResponses(resList)

    PrintResults(options, len(reqList), successes)

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

/**
 * Given a list of requests to make, perform them in parallel and return
 *     once all results are in.
 *
 * Also handles rate limiting for QPX API.
 */
func MakeParallelQPXRequests(reqList []FlightsRequest, config AppConfig) (
    resList []FlightsResult) {

    c := make(chan FlightsResult, len(reqList))
    processed := 0
    limiter := time.Tick(time.Millisecond * 200)

    for _,req := range reqList {
        <-limiter  // Don't overload QPX
        go ParallelQPXRequestHandler(req, config, c)
    }

    for i := 0; i < len(reqList); i++ {
        res := <-c
        resList = append(resList, res)
        processed++
        fmt.Printf("Received %d Out of %d Responses\n", processed, len(reqList))
    }

    return

}

/**
 * Create a request given a set of parameters, make it, and convert it to a more
 *     usable result.
 */
func ParallelQPXRequestHandler(req FlightsRequest, config AppConfig, 
    c chan FlightsResult) {

    qpxReq := BuildQPXRequest(req)
    qpxRes, success := MakeQPXRequest(qpxReq, config)
    res := InterpretQPXResult(qpxRes, success)
    c <- res

}



/**
 * Transform a list of objects containing lists of flight options to just one 
 *     list of flight options.
 */
func FlattenResponses(resList []FlightsResult) (
    optionsList FlightsResultOptionList, successes int) {

    for _,result := range resList {
        if result.Success {
            successes++
            optionsList = append(optionsList, result.Options...)
        }
    }
    sort.Sort(optionsList)
    return

}

