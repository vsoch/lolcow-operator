#!/bin/bash

if [ $# -eq 0 ]; then
    fortune | cowsay | lolcat
else
    echo $@ | cowsay | lolcat
fi

