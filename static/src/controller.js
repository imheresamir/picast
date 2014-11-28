/* globals define */
define(function(require, exports, module) {
    var serverIp = require('serverIp');
    serverIp = 'http://' + serverIp;

    var view = require('main');
    var playlist = view[0];
    var playercontrols = view[1];
    var uiLock = false;

    playercontrols._eventInput.on("prev", function() {
        if(!uiLock) {
            playlist._eventOutput.emit("prevSelect");

            console.log(playlist.currentlySelectedIndex);

            // Talk to server

            playlist._eventOutput.emit("prevPlay");
        }
    });

    playercontrols._eventInput.on("pause", function() {
        console.log("pause");
        //uiLock = true;

        // Talk to server
        var ajax = new XMLHttpRequest();

        ajax.open('POST', serverIp + ':8082/media/pause', true);
        ajax.setRequestHeader('Content-Type', 'application/json');
        ajax.send('{"Pause":"pause"}');
        
        ajax.onreadystatechange = function() {
            if(ajax.readyState == 4 && ajax.status == 200) {
                uiLock = false;
            }
        }

        playercontrols._eventOutput.emit("paused");

    });

    playercontrols._eventInput.on("play", function() {
        console.log("play");
        //uiLock = true;

        // Talk to server
        var ajax = new XMLHttpRequest();

        ajax.open('POST', serverIp + ':8082/media/pause', true);
        ajax.setRequestHeader('Content-Type', 'application/json');
        ajax.send('{"Pause":"play"}');
        
        ajax.onreadystatechange = function() {
            if(ajax.readyState == 4 && ajax.status == 200) {
                uiLock = false;
            }
        }

        playercontrols._eventOutput.emit("playing");
        
    });

    playercontrols._eventInput.on("next", function() {
        if(!uiLock) {
            playlist._eventOutput.emit("nextSelect");

            console.log(playlist.currentlySelectedIndex);

            // Talk to server

            playlist._eventOutput.emit("nextPlay");
        }
    });

    playlist._eventInput.on("itemClicked", function(index) {
        if(!uiLock) {
            playlist._eventOutput.emit("select", index);

            // Talk to server

            playlist._eventOutput.emit("play", index);
        }
    });

});
