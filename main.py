from preferredwaveplayer import *
from datetime import datetime

def main():
    time_string = datetime.now().time()

    [hour, minute, second] = str(time_string).split(":")

    hour_int = int(hour)
    if hour_int > 12 or hour_int == 0:
        mood = "night"
    else:
        mood = "day"

    if hour_int == 0:
        hour_int = 12
    
    print("test")

    play_completely("assets/in-between/the.wav")
    play_completely("assets/mood/{0}.wav".format(mood))
    play_completely("assets/in-between/is.wav")

    play_completely("assets/minutes/halfway-through.wav")
    play_completely("assets/hour/{0}.wav".format(hour_int))
    

def play_completely(file_name):
    sound = playwave(file_name)

    while getIsPlaying(sound):
        continue


if __name__ == "__main__":
    main()