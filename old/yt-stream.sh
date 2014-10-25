#!/bin/bash

youtube-dl -o 'tmp%(autonumber)s' -a batchfile --exec omxplayer
