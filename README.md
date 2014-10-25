pi-cast
=======

##### Broadcast your Chrome browser window and video streams to your RaspberryPi

Pi-cast 0.1 pre-alpha
------------
The goal of this initial release is to make it simple and easy to seamlessly cast Youtube videos from your Chrome browser tab to a secondary display through a RaspberryPi on the local network. The backend core Python server apps are at 80% and the rest has yet to be implemented.

Dependencies/Technologies:
* Backend (rpi)
  * [youtube-dl](https://github.com/rg3/youtube-dl): for downloading/buffering videos to rpi sdcard
  * [omxplayer](https://github.com/popcornmix/omxplayer): playing videos on rpi
  * Python 2.7: on-demand video player daemon (jobserver.py), database reader and video handler (client.py)
    * [Pyro4](https://github.com/irmen/Pyro4): Interprocess communication between jobserver.py and client.py
    * python-daemon
  * sqlite3: playlist database
  * Go 1.0.2: HTTP listen server and database writer
    * [gorilla/mux](https://github.com/gorilla/mux)
* Frontend (chrome)
  * Javascript: Chrome extension
    * [famous-angular](https://github.com/famous/famous-angular): MVC framework with the strength of [Angular](https://github.com/angular/angular.js) and the shine of [Famo.us](https://github.com/famous/famous) - generates our Views and talks to our HTTP server
* Future Frontend (Android/iOS)
  *  [famous-angular](https://github.com/famous/famous-angular) + WebView/CocoonJS/Crosswalk

Pipe Dreams:
* Encode realtime video stream of browser tab using ffmpeg-PNaCL in Chrome and cast it to rpi 
