from math import *
from datetime import datetime
from preferredwaveplayer import *
import wave
OUTFILE = "current_time.temp.wav"
import os
import sys

# todo: command line arguments to use slow assets
def main():
	time_string = datetime.now().time()

	[hour, minute, second] = str(time_string).split(":")

	# [hour, minute, second] = [1, 18, 1]
	(precedence, new_hour, new_minute) = precedence_hour_minutes(hour, minute)
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

	append_to_outfile(wav_files)
	print_speech(wav_files)
	play_completely(OUTFILE)
	
def hour_and_mood(hour_int):
	if hour_int > 12 or hour_int == 0:
		mood = "night"
	else:
		mood = "day"
	
	if hour_int == 0 or hour_int == 24:
		hour_int = 12
	else:
		hour_int %= 12

	return (hour_int, mood)

def precedence_hour_minutes(hour_raw, minutes_raw):
	approximate = float(minutes_raw) % 5
	
	increment = 0
	if approximate > 2:
		increment += 1

	chunk = floor(float(minutes_raw) / 5)
	approx_chunk = (chunk + increment) * 5
	
	new_hour = int(hour_raw)
	
	if approx_chunk == 0:
		return (None, new_hour, "around")
	
	if approx_chunk == 30:
		return (None, new_hour, "halfway-through")
	
	if approx_chunk == 60:
		new_hour += 1
		return (None, new_hour, "around")

	if approx_chunk > 30:
		new_hour += 1
		return ("before", new_hour, str(30 - (approx_chunk % 30)))

	return ("after", new_hour, str(approx_chunk))

def play_completely(file_name):
	sound = playwave(file_name)

	while getIsPlaying(sound):
		continue

# https://stackoverflow.com/questions/61499350/combine-audio-files-in-python
def append_to_outfile(wav_files):
	global OUTFILE

	combined_data = []
	for wav_file in wav_files:
		if not os.path.isfile(wav_file):
			print("({0}): ** Fatal ** {1}/{2} is not a file."
				.format("approxima", os.getcwd(), wav_file))
			sys.exit(1)

		w = wave.open(wav_file, 'rb')
		combined_data.append( [w.getparams(), w.readframes(w.getnframes())] )
		w.close()

	output = wave.open(OUTFILE, 'wb')
	output.setparams(combined_data[0][0])

	for data_point in combined_data:
		output.writeframes(data_point[1])
	output.close()

def print_speech(wav_file_paths):
	for wav_file in wav_file_paths:
		speech = wav_file \
			.split("/")[-1] \
			.split(".")[0] \
		
		if speech.startswith("-"):
			speech = speech.split("-")[-1]
		else:
			speech = speech.replace("-", "\n")
		print(speech)

if __name__ == "__main__":
	main()