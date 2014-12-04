import QtQuick 2.2
import QtQuick.Controls 1.1

ApplicationWindow {
    id: window
    visible: true
    width: 1920
    height: 1080
    color: "#000000"
    opacity: 1

    Image {
        id: art
        width: 640
        height: 640
        transformOrigin: Item.TopLeft
        anchors.topMargin: window.height / 4
        anchors.top: parent.top
        anchors.leftMargin: window.width / 4
        anchors.left: parent.left
        scale: 0.5
        opacity: 1
        z: 1
        fillMode: Image.PreserveAspectFit
        source: currentTrack.artPath
    }
}
