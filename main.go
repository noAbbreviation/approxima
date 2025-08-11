package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	beep "github.com/gopxl/beep/v2"
	"github.com/gopxl/beep/v2/speaker"
	"github.com/gopxl/beep/v2/wav"
)

var assetsFolder string

type Approx struct {
	hour          int
	minuteOffset  int
	isAM          bool
	beforeHalfway bool
}

type Asset struct {
	category, fileName string
}

func main() {
	shortFlag := flag.Bool("short", false, "Shorten the prompt")
	assetsFolderArgs := flag.String("assets", "./assets", "Folder to show assets")
	flag.Parse()

	assetsFolder = *assetsFolderArgs
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
	hour, minute, _ := roundedTime.Clock()

	beforeHalfway := minute < 30
	minuteOffset := func() int {
		if minute < 30 {
			return minute
		}

		if minute == 30 {
			return 30
		}

		hour += 1
		return 60 - minute
	}()

	isAM := hour < 12 || (hour == 12 && !beforeHalfway)
	normalizedHour := func() int {
		if hour%12 == 0 {
			return 12
		}

		return hour % 12
	}()

	approxInstance := Approx{
		hour:          normalizedHour,
		minuteOffset:  minuteOffset,
		isAM:          isAM,
		beforeHalfway: beforeHalfway,
	}

	theSound := Asset{"in-between", "the"}
	moodSound := Asset{
		"mood",
		func() string {
			var moodVariants map[bool]string

			if *shortFlag {
				moodVariants = map[bool]string{
					true:  "am",
					false: "pm",
				}
			} else {
				moodVariants = map[bool]string{
					true:  "day",
					false: "night",
				}
			}

			if approxInstance.isAM {
				return moodVariants[true]
			}

			return moodVariants[false]
		}(),
	}
	isSound := Asset{"in-between", "is"}
	minuteValueSound := Asset{
		"minutes",
		func() string {
			minuteOffset := approxInstance.minuteOffset

			if minuteOffset == 0 {
				return "around"
			} else if minuteOffset == 30 {
				return "halfway-through"
			}

			return fmt.Sprint(minuteOffset)
		}(),
	}
	minuteNameSound := func() Asset {
		minuteOffset := approxInstance.minuteOffset

		if minuteOffset == 0 || minuteOffset == 30 {
			return Asset{}
		}

		return Asset{"minutes", "-connect-minutes"}
	}()
	precedenceSound := func() Asset {
		minuteOffset := approxInstance.minuteOffset

		if minuteOffset == 0 || minuteOffset == 30 {
			return Asset{}
		}

		return Asset{
			"precedence",
			func() string {
				if approxInstance.beforeHalfway {
					return "after"
				}

				return "before"
			}(),
		}
	}()
	hourSound := Asset{"hour", fmt.Sprint(approxInstance.hour)}

	audioAssets := []Asset{
		theSound,
		moodSound,
		isSound,
		minuteValueSound,
		minuteNameSound,
		precedenceSound,
		hourSound,
	}
	if *shortFlag {
		audioAssets = []Asset{
			minuteValueSound,
			precedenceSound,
			hourSound,
			moodSound,
		}
	}

	prompt := []string{}
	for _, asset := range audioAssets {
		if asset.category == "" {
			continue
		}

		promptItem := asset.fileName
		if promptItem[0] == '-' {
			extractedPromptItem := strings.Split(promptItem, "-")
			promptItem = extractedPromptItem[len(extractedPromptItem)-1]
		}

		prompt = append(prompt, promptItem)
	}
	fmt.Println(strings.Join(prompt, " "))

	audioStreamers := []beep.StreamCloser{}
	var audioFormat beep.Format
	defer func() {
		for _, streamer := range audioStreamers {
			streamer.Close()
		}
	}()

	for _, audioAsset := range audioAssets {
		if audioAsset.category == "" {
			continue
		}

		audioFileName := fileNameFromAsset(audioAsset)
		audioFile, err := os.Open(audioFileName)
		if err != nil {
			log.Fatal(err)
		}

		streamer, format, err := wav.Decode(audioFile)
		if err != nil {
			log.Fatal(err)
		}
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

func fileNameFromAsset(asset Asset) string {
	return fmt.Sprintf("%s/%s/%s.wav", assetsFolder, asset.category, asset.fileName)
}
