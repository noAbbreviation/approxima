from math import *
from datetime import datetime
from preferredwaveplayer import *
import wave
import os
import sys

OUTFILE = "current_time.temp.wav"
SCRIPT_NAME = "approxima"
DEFAULT_PLAY_SOUND = True
DEFAULT_PRINT_TEXT = True

PLAY_SOUND = DEFAULT_PLAY_SOUND 
PRINT_TEXT = DEFAULT_PRINT_TEXT
def main():
	set_global_flags(sys.argv[1:])

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
	if PRINT_TEXT:
		print_speech(wav_files)
	if PLAY_SOUND:
		play_completely(OUTFILE)

def set_global_flags(arguments):
	global SCRIPT_NAME
	global PLAY_SOUND
	global PRINT_TEXT

	sound = DEFAULT_PLAY_SOUND
	to_print = DEFAULT_PRINT_TEXT

	for arg in arguments:
		if not arg.startswith("--"):
			print("({0}): ** Fatal ** \"{1}\" is not a valid argument.".format(SCRIPT_NAME, arg))
			sys.exit(1)
		flag = arg[2:]		

		if flag == "sound":
			sound = True
		elif flag == "no-sound":
			sound = False
		elif flag == "print":
			to_print = True
		elif flag == "no-print":
			to_print = False

		elif flag == "help":
			usage_file_path = "usage.txt"
			if not os.path.isfile(usage_file_path):
				print(
					"""({0}): ** Fatal ** {1}/{2} is not available.
					 Please check the original repo for the file."""
						.format(SCRIPT_NAME, os.getcwd(), usage_file_path)
						.replace("\t", "")
						.replace("\n", "")
				)
				sys.exit(1)
			
			file = open(usage_file_path, "r")
			file_contents = file.read()
			print(file_contents)

			file.close()
			sys.exit(0)

		else:
			print("({0}): ** Fatal ** \"{1}\" is not a valid argument.".format(SCRIPT_NAME, arg))
			sys.exit(1)

	PLAY_SOUND = sound
	PRINT_TEXT = to_print

def hour_and_mood(hour_int):
	if hour_int > 12 or hour_int == 0:
		mood = "night"
	else:
		mood = "day"
	
	if hour_int % 12 == 0:
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
	global SCRIPT_NAME

	combined_data = []
	for wav_file in wav_files:
		if not os.path.isfile(wav_file):
			print("({0}): ** Fatal ** {1}/{2} is not a file."
				.format(SCRIPT_NAME, os.getcwd(), wav_file))
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