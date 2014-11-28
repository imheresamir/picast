/* globals define */
define(function(require, exports, module) {
    /********************* Dependencies *********************/
    var Engine = require('famous/core/Engine');
    var Modifier = require('famous/core/Modifier');
    var Transform = require('famous/core/Transform');
    var InputSurface = require('famous/surfaces/InputSurface');
    var Surface = require('famous/core/Surface');
    var FlexibleLayout = require('famous/views/FlexibleLayout');
    var RenderNode = require('famous/core/RenderNode');

    var PlaylistView = require('views/PlaylistView');
    var PlayerControlsView = require('views/PlayerControlsView');
    require('famous/inputs/DesktopEmulationMode');

    /********************* Views and Arrays *********************/
    var mainContext = Engine.createContext();

    var surfaces = [];
    var middleBarItems = [];

    var layout = new FlexibleLayout({
        direction: 1,
        ratios: [0.75, 5.25, 1]
    });

    var middleBarView = new FlexibleLayout({
        direction: 0,
        ratios: [0.35, 0.65]
    });

    var playlist = new PlaylistView({
        sizeFunction: function() {
            return [window.innerWidth * 0.65, window.innerHeight * 5.25 / 7 / 5 -0.5];
        }
    });

    var playercontrols = new PlayerControlsView();

    /********************* Main *********************/

    var background = new Surface({
        properties: {
            backgroundColor: 'rgba(0, 0, 0, 0.8)',
            zIndex: -1
        }
    });

    // Middle Left Bar
    middleBarItems.push(new Surface({
        properties: {
            boxShadow: '-2px -1px 5px black inset'
            //borderRight: '1px solid black',
            //borderTop: 'none'
        }
    }))

    middleBarItems.push(playlist);
    middleBarView.sequenceFrom(middleBarItems);

    // Top Bar
    surfaces.push(new Surface({
        properties: {
            //border: '1px solid black',
            boxShadow: '0px -2px 5px black inset',
            borderBottom: 'none',
            zIndex: -1
        }
    }));
    surfaces.push(middleBarView);
    surfaces.push(playercontrols);

    //setupInputBar();

    layout.sequenceFrom(surfaces);
    //mainContext.setPerspective(1000);
    mainContext.add(background);
    mainContext.add(layout);

    /********************* Setup Functions *********************/

    function setupInputBar() {

        var inputnode = new RenderNode();
        var inputMod = new Modifier({
            align: [0.5, 0.5],
            origin: [0.5, 0.5]
        });
        var input = new InputSurface({
            //size: [50, 50],
            type: 'search',
            outline: 'none',
            properties: {
                //backgroundColor: 'black'
                borderColor: 'rgba(255, 255, 255, 0.3)',
                borderRadius: '30px',
                backgroundColor: 'rgba(0, 0, 0, 0)',
                color: 'white'
                //borderColor: 'rgba(6, 148, 147, 237)'
            }
        });

        inputMod.sizeFrom(function() {
            return [mainContext._size[0] * 7/8, mainContext._size[1] / 12];
        });

        inputnode.add(inputMod).add(input);

        input.on('focus', function() {
            input.setProperties({
                boxShadow: '0 0 5px rgba(6, 148, 147, 237)',
                borderColor: 'rgba(255, 255, 255, 0)',
                outline: 'none',
                backgroundColor: 'rgba(255, 255, 255, 0.2)',
                borderRadius: '30px',
                color: 'white'
            });
        });

        input.on('blur', function() {
            input.setProperties({
                borderColor: 'rgba(255, 255, 255, 0.3)',
                boxShadow: 'none',
                borderRadius: '30px',
                backgroundColor: 'rgba(0, 0, 0, 0)',
                color: 'white'
                //borderColor: 'rgba(6, 148, 147, 237)'
            });
        });

        surfaces.push(inputnode);
    }

    module.exports = [playlist, playercontrols];
});
