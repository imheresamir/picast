/* globals define */
define(function(require, exports, module) {
    var serverConfig = require('serverConfig');
    var serverIp = 'http://' + serverConfig;
    var evtSource = new EventSource(serverIp + ':8081/events');

    var view = require('main');
    var playlist = view[0];
    var playercontrols = view[1];
    var playLock = true;

    evtSource.addEventListener("playerStateChanged", function(e) {
        console.log("Server event: playerState - " + e.data);
        if(e.data == "stopped") {
            playLock = true;
            playercontrols._eventOutput.emit("stopped");
        } else if(e.data == "paused") {
            playLock = false;
            playercontrols._eventOutput.emit("paused");
        } else if(e.data == "playing") {
            playLock = false;
            playercontrols._eventOutput.emit("playing");
        }
    });

    playercontrols._eventInput.on("pause", function() {
        if(!playLock) {
            console.log("pause");
            //uiLock = true;

            // Talk to server
            var ajax = new XMLHttpRequest();

            ajax.open('GET', serverIp + ':8082/media/pause', true);
            ajax.setRequestHeader('Content-Type', 'application/json');
            ajax.send('{"Pause":"pause"}');
            
            ajax.onreadystatechange = function() {
                if(ajax.readyState == 4 && ajax.status == 200) {
                    uiLock = false;
                }
            }

            //playercontrols._eventOutput.emit("paused");
        }

    });

    playercontrols._eventInput.on("play", function() {
        if(!playLock) {
            console.log("play");
            //uiLock = true;

            // Talk to server
            var ajax = new XMLHttpRequest();

            ajax.open('GET', serverIp + ':8082/media/pause', true);
            ajax.setRequestHeader('Content-Type', 'application/json');
            ajax.send('{"Pause":"play"}');
            
            ajax.onreadystatechange = function() {
                if(ajax.readyState == 4 && ajax.status == 200) {
                    uiLock = false;
                }
            }

            //playercontrols._eventOutput.emit("playing");
        }
        
    });

    playercontrols._eventInput.on("prev", function() {
        playlist._eventOutput.emit("prevSelect");

        console.log(playlist.currentlySelectedIndex);

        // Talk to server

        playlist._eventOutput.emit("prevPlay");
    });

    playercontrols._eventInput.on("next", function() {
        playlist._eventOutput.emit("nextSelect");

        console.log(playlist.currentlySelectedIndex);

        // Talk to server

        playlist._eventOutput.emit("nextPlay");
    });

    playlist._eventInput.on("itemClicked", function(index) {
        playlist._eventOutput.emit("select", index);

        // Talk to server

        playlist._eventOutput.emit("play", index);
    });

});
