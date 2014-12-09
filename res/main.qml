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

    Text {
        id: titleText
        y: 398
        width: 445
        height: 65
        color: "#ffffff"
        text: qsTr(currentTrack.title)
        anchors.bottom: art.top
        anchors.bottomMargin: -192
        anchors.left: art.horizontalCenter
        anchors.leftMargin: 55
        verticalAlignment: Text.AlignTop
        horizontalAlignment: Text.AlignLeft
        font.family: "Verdana"
        font.bold: true
        font.pixelSize: 50
    }

    Text {
        id: artistText
        width: 445
        height: 40
        color: "#ffffff"
        text: qsTr(currentTrack.artist)
        anchors.top: titleText.bottom
        anchors.topMargin: -6
        anchors.left: art.right
        anchors.leftMargin: -265
        verticalAlignment: Text.AlignVCenter
        font.pixelSize: 25
        font.bold: false
        font.family: "Verdana"
        horizontalAlignment: Text.AlignLeft
    }

    Text {
        id: albumText
        width: 445
        height: 40
        color: "#ffffff"
        text: qsTr(currentTrack.album)
        anchors.top: artistText.bottom
        anchors.topMargin: -6
        anchors.left: art.right
        anchors.leftMargin: -265
        verticalAlignment: Text.AlignVCenter
        font.pixelSize: 25
        font.bold: false
        font.family: "Verdana"
        horizontalAlignment: Text.AlignLeft
    }
}
