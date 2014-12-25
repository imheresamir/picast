/* globals define */
define(function(require, exports, module) {
    var serverConfig = require('serverConfig');
    var serverIp = 'http://' + serverConfig;
    var evtSource = new EventSource(serverIp + ':8081/events');

    var view = require('main');
    var playlist = view[0];
    var playercontrols = view[1];
    var playLock = true;

    // Get initial player state
    var ajax = new XMLHttpRequest();

    ajax.open('GET', serverIp + ':8082/media/status', true);
    ajax.setRequestHeader('Content-Type', 'application/json');
    ajax.send();

    ajax.onreadystatechange = function() {
        if(ajax.readyState == 4 && ajax.status == 200) {
            if(ajax.responseText.indexOf('playing') != -1) {
                playLock = false;
                playercontrols._eventOutput.emit("playing");
            } else if(ajax.responseText.indexOf('paused') != -1) {
                playLock = false;
                playercontrols._eventOutput.emit("paused");
            } else {
                playLock = true;
                playercontrols._eventOutput.emit("stopped");
            }
        }
    }

    var ajax2 = new XMLHttpRequest();

    ajax2.open('GET', serverIp + ':8082/media/playlist', true);
    ajax2.setRequestHeader('Content-Type', 'application/json');
    ajax2.send();

    ajax2.onreadystatechange = function() {
        if(ajax2.readyState == 4 && ajax2.status == 200) {
            var jsonData = JSON.parse(ajax2.responseText);
            if(jsonData.Playlist) {

                for(var i = 0; i < jsonData.Playlist.length; i++) {
                    playlist.AddEntry(jsonData.Playlist[i]);
                }
                playlist.Reconstruct();

                playLock = false;
                playercontrols._eventOutput.emit("paused");

            }
        }
    }

    evtSource.addEventListener("playerStateChanged", function(e) {
        console.log("Server event: playerState - " + e.data);
        if(e.data == "stopped") {
            //playLock = true;
            playercontrols._eventOutput.emit("stopped");
        } else if(e.data == "paused") {
            playLock = false;
            playercontrols._eventOutput.emit("paused");
        } else if(e.data == "playing") {
            playLock = false;
            playercontrols._eventOutput.emit("playing");
        }
    });

    evtSource.addEventListener("playlistChanged", function(e) {
        console.log("Server event: playlistChanged - " + e.data);
        
        var ajax3 = new XMLHttpRequest();

        ajax3.open('GET', serverIp + ':8082/media/playlist', true);
        ajax3.setRequestHeader('Content-Type', 'application/json');
        ajax3.send();

        ajax3.onreadystatechange = function() {
            if(ajax3.readyState == 4 && ajax3.status == 200) {
                var jsonData = JSON.parse(ajax3.responseText);
                if(jsonData.Playlist) {

                    playlist.Clear();

                    for(var i = 0; i < jsonData.Playlist.length; i++) {
                        playlist.AddEntry(jsonData.Playlist[i]);
                    }

                    playlist.Reconstruct();

                }
            }
        }
    });


    playercontrols._eventInput.on("pause", function() {
        if(!playLock) {
            console.log("pause");

            // Talk to server
            var ajax = new XMLHttpRequest();

            ajax.open('GET', serverIp + ':8082/media/pause', true);
            ajax.setRequestHeader('Content-Type', 'application/json');
            ajax.send(null);
            
            playercontrols._eventOutput.emit("paused");
        }

    });

    playercontrols._eventInput.on("play", function() {
        if(!playLock) {
            console.log("play");

            // Talk to server
            var ajax = new XMLHttpRequest();

            ajax.open('GET', serverIp + ':8082/media/play', true);
            ajax.setRequestHeader('Content-Type', 'application/json');
            ajax.send(null);

            playercontrols._eventOutput.emit("playing");
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

    playlist._eventInput.on("playlistEntryAdded", function(data) {
        // Talk to server
        var ajax = new XMLHttpRequest();
        var serverResponse;

        ajax.open('POST', serverIp + ':8082/media/add', true);
        ajax.setRequestHeader('Content-Type', 'application/json');
        ajax.send('{"Url":"' + data + '"}');
        
    });

});
