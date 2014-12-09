#!/bin/bash

#SPOTIFYLIBS=`pwd`/lib/spotify
#SPOTIFYPCDIR=$SPOTIFYLIBS/lib/pkgconfig

#export PKG_CONFIG_PATH=$SPOTIFYPCDIR
go build -o picast -x main/picast.go

