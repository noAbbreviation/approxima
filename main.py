from preferredwaveplayer import *
from datetime import datetime

def main():
    time_string = datetime.now().time()

    [hour, minute, second] = str(time_string).split(":")

    (new_hour, mood) = hour_and_mood(hour)
    (precedence, new_minute) = precedence_and_minutes(minute)

    print("test")

    play_completely("assets/in-between/the.wav")
    play_completely("assets/mood/{0}.wav".format(mood))
    play_completely("assets/in-between/is.wav")

    play_completely("assets/minutes/{0}.wav".format(new_minute))
    
    if precedence != None:
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

    return (hour_int, mood)

def precedence_and_minutes(minutes_raw):
    approx = 60 - (round(1.0 * int(minutes_raw) / 5) * 5)

    if approx == 0:
        return ("before", "around")
    elif approx == 30:
        return (None, "halfway-through")
    elif approx < 30:
        return ("before", str(approx))
    else:
        return ("after", str(approx % 30))


def play_completely(file_name):
    sound = playwave(file_name)

    while getIsPlaying(sound):
        continue


if __name__ == "__main__":
    main()