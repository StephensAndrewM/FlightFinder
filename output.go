package main

import(
	"fmt"
	"bytes"
	"github.com/fatih/color"
)

func Min(x, y int) int {
    if x < y {
        return x
    }
    return y
}

func PrintResults(optionsList []FlightsResultOption) {

	const WIDTH = 50
	costFont 	:= color.New(color.FgYellow, color.Bold)

	for i := 0; i < Min(len(optionsList), 10); i++ {
		option := optionsList[i]

		fmt.Println(repeatChar("=", WIDTH))
		fmt.Printf("Cost:      ")
		costFont.Printf("$%.2f\n", option.Price)

		fmt.Println(repeatChar("-", WIDTH))
		fmt.Printf("Outbound:  ")
		PrintSlice(option.Slices[0])

		fmt.Println(repeatChar("-", WIDTH))
		fmt.Printf("Inbound:   ")
		PrintSlice(option.Slices[1])

	}

}

func PrintSlice(slice FlightsResultSlice) {

	const DATETIME_FMT = "Mon Jan 02 03:04 PM MST"
	flightMainFont 		:= color.New(color.FgCyan, color.Bold)
	flightDetailFont 	:= color.New(color.FgCyan)
	warningFont 		:= color.New(color.FgRed, color.Bold)

	if len(slice.Segments) > 1 {
		fmt.Printf("\n")
	}

	for _,segment := range slice.Segments {

		flightMainFont.Printf("%s -> %s\n", segment.Origin, segment.Destination)

		fmt.Printf("Flight:    ")
		flightDetailFont.Printf("%s (%s)\n", segment.FlightNumber, segment.Airline)
		fmt.Printf("Departure: ")
		flightDetailFont.Printf("%s\n", segment.DepartureTime.Format(DATETIME_FMT))
		fmt.Printf("Arrival:   ")
		flightDetailFont.Printf("%s\n", segment.ArrivalTime.Format(DATETIME_FMT))

		if segment.NumLegs > 1 {
			warningFont.Printf("Multiple Legs: %d\n", segment.NumLegs)
		}
	}

}

func repeatChar(char string, num int) string {
	var buffer bytes.Buffer
	for i := 0; i < num; i++ {
		buffer.WriteString(char)
	}
	return buffer.String()
}