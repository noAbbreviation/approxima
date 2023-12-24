from math import *
from datetime import datetime
from preferredwaveplayer import *
import wave
OUTFILE = "current_time.temp.wav"

# todo: tell afternoon, midnight times
def main():
    time_string = datetime.now().time()

    [hour, minute, second] = str(time_string).split(":")

    # [hour, minute, second] = [1, 18, 1]
    (precedence, new_minute) = precedence_and_minutes(minute)

    if precedence == "before":
        new_hour = int(hour) + 1
    else:
        new_hour = int(hour)

    (new_hour, mood) = hour_and_mood(new_hour)

    wav_files = [
        "assets/in-between/the.wav",
        "assets/mood/{0}.wav".format(mood),
        "assets/in-between/is.wav",
        "assets/minutes/{0}.wav".format(new_minute)
    ]
    if precedence != None:
        wav_files += [
            "assets/minutes/-connect-minutes.wav",
            "assets/precedence/{0}.wav".format(precedence)
        ]
    wav_files += ["assets/hour/{0}.wav".format(new_hour)]

    append_wav_files(wav_files)
    play_completely(OUTFILE)
    
def hour_and_mood(hour_raw):
    hour_int = int(hour_raw)
    if hour_int > 12 or hour_int == 0:
        mood = "night"
    else:
        mood = "day"
    
    if hour_int == 0 or hour_int == 24:
        hour_int = 12
    else:
        hour_int %= 12

    return (hour_int, mood)

def precedence_and_minutes(minutes_raw):
    approximate = float(minutes_raw) % 5
    if approximate > 2:
        increment = 1
    else:
        increment = 0
    chunk = floor(float(minutes_raw) / 5)
    approx_chunk = (chunk + increment) * 5
    print(approx_chunk)

    if approx_chunk == 0:
        return (None, "around")
    elif approx_chunk == 30:
        return (None, "halfway-through")
    elif approx_chunk > 30:
        return ("before", str(30 - (approx_chunk % 30)))
    else:
        return ("after", str(approx_chunk))

def play_completely(file_name):
    sound = playwave(file_name)

    while getIsPlaying(sound):
        continue

# https://stackoverflow.com/questions/61499350/combine-audio-files-in-python
def append_wav_files(wav_files):
    global OUTFILE

    combined_data = []
    print(wav_files)
    for wav_file in wav_files:
        w = wave.open(wav_file, 'rb')
        combined_data.append( [w.getparams(), w.readframes(w.getnframes())] )
        w.close()

    output = wave.open(OUTFILE, 'wb')
    output.setparams(combined_data[0][0])

    for data_point in combined_data:
        output.writeframes(data_point[1])
    output.close()

if __name__ == "__main__":
    main()