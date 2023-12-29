# Approxima

A command line program to approximately tell time (in chunks of 5 minutes) using Python.

#### Currently has:
- Windows executable (Call using cmd does not work)
- Linux executable (setup needed, steps below)

## To run:

### Python
Setup venv dependencies:
```
$ python3 -m venv env
  env/bin/pip3 install <lines from venv-installs.txt> <add preferredwaveplayer as a safe measure>
```

Now, run:
```
$ python main.py
```

### Windows
```
$ approxima.exe
```

### Linux (Ubuntu)
Setup:
Remove `_internals` folder and replace with `_linux-internals`
```bash
$ rm -rf _internals
  mkdir _internals
  cp -r _linux-internals/* _internals
```

Now, run:
```bash
$ ./main
```
Original audio files generated from [ttsmaker](https://ttsmaker.com/).
