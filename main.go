package main

import (
	"embed"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"strconv"
	"strings"
	"time"

	beep "github.com/gopxl/beep/v2"
	"github.com/gopxl/beep/v2/speaker"
	"github.com/gopxl/beep/v2/wav"
)

var (
	//go:embed assets/*
	embeddedAssets embed.FS

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

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  approxima-go [OPTION]... [<HH:MM:SS>]")

	fmt.Println("\nPipe UNIX time as substitute for command line arguments.")

	fmt.Println("\nFlags: ")
	flag.PrintDefaults()
}

func main() {
	_false := false
	helpFlag := &_false

	flag.BoolVar(helpFlag, "h", false, "Display this help then exit.")
	flag.BoolVar(helpFlag, "help", false, "Display this help then exit.")

	shortFlag := flag.Bool("short", false, "Use the shorter prompt format.")
	assetsFolderFlag := flag.String("assets", "", "Asset folder to use.")
	silentFlag := flag.Bool("silent", false, "Only print the prompt then exit.")

	flag.Parse()

	if *helpFlag {
		printUsage()
		os.Exit(0)
	}

	defaultTime := time.Now()
	if customTimeArg := flag.Arg(0); len(customTimeArg) != 0 {
		timeFmt := "15:04:05"

		customTime, err := time.Parse(timeFmt, customTimeArg)
		if err != nil {
			fmt.Println("Error parsing the custom time: Accepted format is <HH:MM:SS>")
			os.Exit(1)
		}

		defaultTime = customTime
	}

	timeData, err := checkPipedStdinForTime(defaultTime)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fiveMinutes, err := time.ParseDuration("5m")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	roundedTime := timeData.Round(fiveMinutes)
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

		if strings.HasPrefix(promptItem, "-") {
			extractedPromptItem := strings.Split(promptItem, "-")
			promptItem = extractedPromptItem[len(extractedPromptItem)-1]
		}

		if strings.Contains(promptItem, "-") {
			promptItem = strings.ReplaceAll(promptItem, "-", " ")
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

	assetsFolderName := "assets"
	assetFS := fs.FS(embeddedAssets)

	if assetsFolderArg := *assetsFolderFlag; len(assetsFolderArg) != 0 {
		assetsFolderName = assetsFolderArg
		assetFS = os.DirFS(".")
	}

	for _, audioAsset := range audioAssets {
		if audioAsset.category == "" {
			continue
		}

		audioFileName := fmt.Sprintf(
			"%s/%s/%s.wav", assetsFolderName, audioAsset.category, audioAsset.fileName,
		)
		audioFile, err := assetFS.Open(audioFileName)
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

func checkPipedStdinForTime(defaultTime time.Time) (time.Time, error) {
	fileStat, err := os.Stdin.Stat()
	if err != nil {
		fmt.Println("error reading stardard input:", err)
		return time.Time{}, err
	}

	noPipedStdIn := fileStat.Mode()&os.ModeNamedPipe == 0
	if noPipedStdIn {
		return defaultTime, nil
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
