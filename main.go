package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	beep "github.com/gopxl/beep/v2"
	"github.com/gopxl/beep/v2/speaker"
	"github.com/gopxl/beep/v2/wav"
)

var (
	assetsFolder string

	InvalidTimeFormatE = errors.New("Format should be Unix Time(in seconds).")
)

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
	silentFlag := flag.Bool("silent", false, "Only print the prompt then exit")
	flag.Parse()

	assetsFolder = *assetsFolderArgs
	testing := false

	currentTime := time.Now()

	if testing {
		timeFmt := "15:04:05"

		mockTime, err := time.Parse(timeFmt, "11:58:29")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		currentTime = mockTime
	}

	timeToProcess, err := checkPipedStdinForTime(currentTime)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fiveMinutes, err := time.ParseDuration("5m")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	roundedTime := timeToProcess.Round(fiveMinutes)
	hour, minute, _ := roundedTime.Clock()

	beforeHalfway := minute < 30
	minuteOffset := func() int {
		if minute < 30 {
			return minute
		}

		if minute == 30 {
			return 30
		}

		hour = (hour + 1) % 24
		return 60 - minute
	}()

	isAM := hour < 12 || (hour == 12 && !beforeHalfway)
	normalizedHour := func() int {
		hour %= 12

		if hour == 0 {
			return 12
		}

		return hour
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

			return moodVariants[approxInstance.isAM]
		}(),
	}
	isSound := Asset{"in-between", "is"}
	minuteValueSound := Asset{
		"minutes",
		func() string {
			minuteOffset := approxInstance.minuteOffset

			if minuteOffset == 0 {
				return "around"
			}

			if minuteOffset == 30 {
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

	if *silentFlag {
		os.Exit(0)
	}

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

		audioFileName := fmt.Sprintf(
			"%s/%s/%s.wav", assetsFolder, audioAsset.category, audioAsset.fileName,
		)
		audioFile, err := os.Open(audioFileName)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		streamer, format, err := wav.Decode(audioFile)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		audioFormat = format

		audioStreamers = append(audioStreamers, streamer)
	}

	var combinedStream beep.Streamer = audioStreamers[0]
	for _, audioStream := range audioStreamers[1:] {
		combinedStream = beep.Seq(combinedStream, audioStream)
	}

	speakerDone := make(chan bool)
	timeout := time.NewTimer(time.Second * 10)

	speaker.Init(audioFormat.SampleRate, audioFormat.SampleRate.N(time.Second/10))
	speaker.Play(beep.Seq(combinedStream, beep.Callback(func() {
		speakerDone <- true
	})))

	select {
	case <-timeout.C:
		fmt.Println("Playback exceeded ten(10) seconds.")
		os.Exit(1)
	case <-speakerDone:
		return
	}
}

func checkPipedStdinForTime(currentTime time.Time) (time.Time, error) {
	fileStat, err := os.Stdin.Stat()
	if err != nil {
		fmt.Println("error reading stardard input:", err)
		return time.Time{}, err
	}

	noPipedStdIn := fileStat.Mode()&os.ModeNamedPipe == 0
	if noPipedStdIn {
		return currentTime, nil
	}

	buffer := make([]byte, 32)
	bufLen, err := os.Stdin.Read(buffer)

	if err != nil && err != io.EOF {
		return time.Time{}, fmt.Errorf("Reading standard input failed: %v", err)
	}

	pipedTime := strings.TrimSpace(string(buffer[:bufLen]))
	seconds, err := strconv.Atoi(pipedTime)

	if err != nil {
		return time.Time{}, InvalidTimeFormatE
	}

	return time.Unix(int64(seconds), 0), nil
}
