package main

/*var Input = InputParams{
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
    DryRun: true,
    CacheOK: true,
}*/

/*var Input = InputParams{
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

var Input = InputParams{
    OriginAirport: "SFO",
    DestAirports: []string{"ORD"},
    Outbound: DirectionParams{
        Date: "2017-03-29",
    },
    Inbound: DirectionParams{
        Date: "2017-03-30",
    },
    NumPassengers: 1,
    DryRun: false,
    CacheOK: true,
}