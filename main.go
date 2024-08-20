package main

import (
	"fmt"
	"log"
	"time"
)

type Approx struct {
	hour          int
	minuteOffset  int
	isAM          bool
	beforeHalfway bool
}

func main() {
	currentTime := time.Now()
	testing := false

	if testing {
		timeFmt := "03:04:05"
		mockTime, err := time.Parse(timeFmt, "12:22:29")
		if err != nil {
			log.Fatal(err)
		}
		currentTime = mockTime
	}

	fiveMinutes, err := time.ParseDuration("5m")
	if err != nil {
		log.Fatal(err)
	}

	roundedTime := currentTime.Round(fiveMinutes)
	fmt.Println("roundedTime:", roundedTime)

	hour, minute, _ := roundedTime.Clock()
	minuteOffset := (func() int {
		if minute < 30 {
			return minute
		}

		if minute == 30 {
			return 0
		}

		return 60 - minute
	})()
	isAM := hour < 12
	beforeHalfway := minute < 30

	approxInstance := Approx{
		hour:          hour,
		minuteOffset:  minuteOffset,
		isAM:          isAM,
		beforeHalfway: beforeHalfway,
	}
	fmt.Printf("approx: %+v\n", approxInstance)
}
