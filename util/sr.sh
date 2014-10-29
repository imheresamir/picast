#!/bin/bash

# Get proxy port
PORT=`curl -s -X GET localhost:9090/proxy | grep -Po '{"port":\K[^}]*'`

# Download HAR
curl -s -X PUT -d "captureHeaders=true&captureContent=true&captureBinaryContent=true" http://localhost:9090/proxy/$PORT/har > data.har 

# Extract video data
mkdir -p store
../bin/harx-rpi -xmd "video/MP2T" store/ data.har

# Cat video data
cat store/* store/video.ts

# Play test
omxplayer store/video.ts
