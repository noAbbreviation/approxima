from math import *
from preferredwaveplayer import *
from datetime import datetime

# todo: tell afternoon, midnight times
# todo: combine to a tmp wav file, then play all at once
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

    play_completely("assets/in-between/the.wav")
    play_completely("assets/mood/{0}.wav".format(mood))
    play_completely("assets/in-between/is.wav")

    play_completely("assets/minutes/{0}.wav".format(new_minute))
    
    if precedence != None and new_minute != "around":
        play_completely("assets/minutes/-connect-minutes.wav")
        play_completely("assets/precedence/{0}.wav".format(precedence))

    play_completely("assets/hour/{0}.wav".format(new_hour))
    
def hour_and_mood(hour_raw):
    hour_int = int(hour_raw)
    if hour_int > 12 or hour_int == 0:
        mood = "night"
    else:
        mood = "day"
    
    if hour_int == 0:
        hour_int = 12

    return (hour_int % 12, mood)

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
        return ("before", "around")
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


if __name__ == "__main__":
    main()