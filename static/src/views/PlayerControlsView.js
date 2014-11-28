define(function(require, exports, module) {
    var View = require('famous/core/View');
    var FlexibleLayout = require('famous/views/FlexibleLayout');
    var Modifier = require('famous/core/Modifier');
    var RenderNode = require('famous/core/RenderNode');
    var Surface = require('famous/core/Surface');
    var ImageSurface = require('famous/surfaces/ImageSurface');
    var EventHandler = require('famous/core/EventHandler');

    var bottomBarItems = [];
    var playerControlItems = [];
    var eventHandler = new EventHandler();

    function PlayerControlsView() {
        View.apply(this, arguments);

        eventHandler.pipe(this);

        // bottomBar holds playerControls
        this.bottomBar = new FlexibleLayout({
            direction: 0,
            ratios: [1, 2, 1]
        });

        this.playerControls = new FlexibleLayout({
            direction: 0,
            ratios: [1, 1.3, 1]
        });

        // construct player controls and add click and touch listeners
        _construct.call(this);

        this.playerControls.sequenceFrom(playerControlItems);

        bottomBarItems.push(new Surface({
            properties: {
                //backgroundColor: 'rgba(0, 0, 0, 0.9)',
                //border: '1px solid black',
                //borderTopRightRadius: '8px',
                //borderBottomRightRadius: '8px',
            }

        }));

        bottomBarItems.push(this.playerControls);

        bottomBarItems.push(new Surface({
            properties: {
                //backgroundColor: 'rgba(0, 0, 0, 0.9)',
                //border: '1px solid black',
                //borderTopRightRadius: '8px',
                //borderBottomRightRadius: '8px',
            }

        }));

        this.bottomBar.sequenceFrom(bottomBarItems);

        this.add(new Surface({
            properties: {
                //border: '1px solid black'
                boxShadow: '0 -2px 5px black',
                zIndex: -1
            }
        }));
        this.add(this.bottomBar);

    }

    PlayerControlsView.prototype = Object.create(View.prototype);
    PlayerControlsView.prototype.constructor = PlayerControlsView;

    PlayerControlsView.DEFAULT_OPTIONS = {};

    function _construct() {
        playerControlItems.push(new ImageSurface({
            content: 'assets/prev.svg',
            properties: {
                //backgroundColor: 'rgba(0, 0, 0, 0.9)',
                //border: 'none',
                // borderRadius: '10px'
            }
        }));

        playerControlItems.push(new ImageSurface({
            content: 'assets/pause.svg',
            properties: {
                //backgroundColor: 'rgba(0, 0, 0, 0.9)',
                //border: 'none',
                // borderRadius: '10px'
            }
        }));

        playerControlItems.push(new ImageSurface({
            content: 'assets/next.svg',
            properties: {
                //backgroundColor: 'rgba(0, 0, 0, 0.9)',
                //border: 'none',
                // borderRadius: '10px'
            }
        }));

        playerControlItems[0].on('mousedown', function() {
            playerControlItems[0].setContent('assets/prev_clicked.svg');
        });
        playerControlItems[0].on('mouseup', function() {
            playerControlItems[0].setContent('assets/prev.svg');

            eventHandler.emit("prev");
        });
        playerControlItems[0].on('touchstart', function() {
            playerControlItems[0].setContent('assets/prev_clicked.svg');
        });
        playerControlItems[0].on('touchend', function() {
            playerControlItems[0].setContent('assets/prev.svg');

            eventHandler.emit("prev");
        });

        playerControlItems[1].on('mousedown', function() {
            if (playerControlItems[1]._imageUrl.indexOf('play') == -1)
                playerControlItems[1].setContent('assets/pause_clicked.svg');
            else
                playerControlItems[1].setContent('assets/play_clicked.svg');
        });
        playerControlItems[1].on('mouseup', function() {
            if (playerControlItems[1]._imageUrl.indexOf('play') == -1) {
                eventHandler.emit("pause");
            }
            else {
                eventHandler.emit("play");
            }

        });
        this.on('paused', function() {
            playerControlItems[1].setContent('assets/play.svg');
        });
        this.on('playing', function() {
            playerControlItems[1].setContent('assets/pause.svg');
        });

        playerControlItems[1].on('touchstart', function() {
            if (playerControlItems[1]._imageUrl.indexOf('play') == -1)
                playerControlItems[1].setContent('assets/pause_clicked.svg');
            else
                playerControlItems[1].setContent('assets/play_clicked.svg');
        });
        playerControlItems[1].on('touchend', function() {
            if (playerControlItems[1]._imageUrl.indexOf('play') == -1) {
                eventHandler.emit("pause");
            }
            else {
                eventHandler.emit("play");
            }
        });

        playerControlItems[2].on('mousedown', function() {
            playerControlItems[2].setContent('assets/next_clicked.svg');
        });
        playerControlItems[2].on('mouseup', function() {
            playerControlItems[2].setContent('assets/next.svg');
            eventHandler.emit("next");
        });
        playerControlItems[2].on('touchstart', function() {
            playerControlItems[2].setContent('assets/next_clicked.svg');
        });
        playerControlItems[2].on('touchend', function() {
            playerControlItems[2].setContent('assets/next.svg');
            eventHandler.emit("next");
        });

    }

    module.exports = PlayerControlsView;
});
