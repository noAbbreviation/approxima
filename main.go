package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gopxl/beep"
	"github.com/gopxl/beep/speaker"
	"github.com/gopxl/beep/wav"
)

const assetFolder = "./assets"

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
		timeFmt := "15:04:05"
		mockTime, err := time.Parse(timeFmt, "11:58:29")
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

	beforeHalfway := minute < 30
	minuteOffset := (func() int {
		if minute < 30 {
			return minute
		}

		if minute == 30 {
			return 30
		}

		hour += 1
		return 60 - minute
	})()

	isAM := hour < 12
	normalizedHour := (func() int {
		if hour == 12 || hour == 0 {
			return 12
		}

		return hour % 12
	})()

	approxInstance := Approx{
		hour:          normalizedHour,
		minuteOffset:  minuteOffset,
		isAM:          isAM,
		beforeHalfway: beforeHalfway,
	}
	fmt.Printf("approx: %+v\n", approxInstance)

	theSound := assetAudioFileName("in-between", "the")
	moodSound := assetAudioFileName("mood", (func() string {
		if approxInstance.isAM {
			return "day"
		}
		return "night"
	})())
	isSound := assetAudioFileName("in-between", "is")
	minuteValueSound := assetAudioFileName("minutes", (func() string {
		minuteOffset := approxInstance.minuteOffset

		if minuteOffset == 0 {
			return "around"
		} else if minuteOffset == 30 {
			return "halfway-through"
		}

		return fmt.Sprint(minuteOffset)
	})())
	minuteNameSound := (func() string {
		minuteOffset := approxInstance.minuteOffset
		if minuteOffset != 0 && minuteOffset != 30 {
			return assetAudioFileName("minutes", "-connect-minutes")
		}
		return ""
	})()
	precedenceSound := (func() string {
		minuteOffset := approxInstance.minuteOffset
		if minuteOffset != 0 && minuteOffset != 30 {
			return assetAudioFileName("precedence", (func() string {
				if approxInstance.beforeHalfway {
					return "after"
				}
				return "before"
			})())
		}
		return ""
	})()
	hourSound := assetAudioFileName("hour", fmt.Sprint(approxInstance.hour))

	audioFileNames := []string{
		theSound,
		moodSound,
		isSound,
		minuteValueSound,
		minuteNameSound,
		precedenceSound,
		hourSound,
	}

	audioStreamers := []beep.StreamCloser{}
	defer func() {
		for _, streamer := range audioStreamers {
			streamer.Close()
		}
	}()

	var audioFormat beep.Format

	for _, audioFileName := range audioFileNames {
		if audioFileName == "" {
			continue
		}

		audioFile, err := os.Open(audioFileName)
		if err != nil {
			log.Fatal(err)
		}

		streamer, format, err := wav.Decode(audioFile)
		audioFormat = format

		audioStreamers = append(audioStreamers, streamer)
	}

	var combinedStream beep.Streamer = audioStreamers[0]
	for _, audioStream := range audioStreamers[1:] {
		combinedStream = beep.Seq(combinedStream, audioStream)
	}

	done := make(chan bool)
	speaker.Init(audioFormat.SampleRate, audioFormat.SampleRate.N(time.Second/10))
	speaker.Play(beep.Seq(combinedStream, beep.Callback(func() {
		done <- true
	})))

	<-done
}

func assetAudioFileName(category, fileName string) string {
	return fmt.Sprintf("%s/%s/%s.wav", assetFolder, category, fileName)
}
