#!/bin/bash

wisdom="$@"
if [ $# -eq 0 ]; then
    wisdom=$(fortune)
fi

# Always show wisdom in terminal with cowsay and colored
echo $wisdom | cowsay | lolcat

# And finish providing to web server to show in GUI
python3 /code/run.py "${wisdom}"

