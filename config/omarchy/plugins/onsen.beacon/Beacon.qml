import QtQuick
import Quickshell
import Quickshell.Io
import Quickshell.Wayland
import Quickshell.Hyprland
import qs.Ui

// Onsen's focus beacon. Hyprland summons it when SUPER goes down and hides it
// when SUPER comes up (config/hypr/bindings.lua, o.bind_hold). It shows only
// after a short hold, so SUPER chords never flash it, then marks the centre of
// the focused window, or of the focused screen when it has none, with a
// breathing core and ripples like a drop on hot water. While SUPER stays down
// it follows focus to wherever a chord moves it. On a terminal holding a
// tmux client it marks the active pane rather than the whole window. Visual only: no keyboard
// focus, and an empty input region, so it never takes a click.
Item {
  id: root

  readonly property color green: "#8FD9A8"
  readonly property int holdDelay: 220      // ms of SUPER before it appears
  readonly property int safetyLimit: 10000  // ms; a lost release never strands it

  property bool held: false
  property bool shown: false
  property real cx: -1000
  property real cy: -1000
  property string monitorName: ""

  function open(payloadJson) {
    held = true
    locate.running = true
    delay.restart()
  }

  function close() {
    held = false
    delay.stop()
    shown = false
  }

  Timer {
    id: delay
    interval: root.holdDelay
    onTriggered: if (root.held) root.shown = true
  }

  Timer {
    interval: root.safetyLimit
    running: root.shown
    onTriggered: root.close()
  }

  // locate.py says where: the centre of the focused window, of the active
  // tmux pane when that window is a terminal holding a tmux client, or of the
  // screen when it has no window, in logical pixels on the focused monitor.
  function place(text) {
    var spot = null
    try { spot = JSON.parse(String(text || "").trim() || "null") } catch (e) {}
    if (!spot) return
    cx = spot.x
    cy = spot.y

    // Focus moved to another screen mid-hold: the overlay maps on the focused
    // screen when it opens, so close and reopen it there.
    if (monitorName !== "" && spot.monitor !== monitorName && shown) {
      shown = false
      Qt.callLater(function() { root.shown = root.held })
    }
    monitorName = spot.monitor
  }

  Process {
    id: locate
    command: ["python3", String(Qt.resolvedUrl("locate.py")).replace(/^file:\/\//, "")]
    stdout: StdioCollector {
      waitForEnd: true
      onStreamFinished: root.place(text)
    }
  }

  Connections {
    target: Hyprland
    function onRawEvent(event) {
      if (!root.held || !event) return
      var name = String(event.name)
      if (name === "activewindowv2" || name === "focusedmonv2" || name === "workspacev2" || name === "movewindowv2")
        locate.running = true
    }
  }

  OverlayWindow {
    id: panel
    shown: root.shown
    shownKeyboardFocus: WlrKeyboardFocus.None
    WlrLayershell.namespace: "onsen-beacon"
    mask: Region {}

    Item {
      id: beacon
      x: root.cx
      y: root.cy
      opacity: root.shown ? 1 : 0
      Behavior on opacity { NumberAnimation { duration: 140 } }
      Behavior on x { NumberAnimation { duration: 180; easing.type: Easing.OutCubic } }
      Behavior on y { NumberAnimation { duration: 180; easing.type: Easing.OutCubic } }

      // Ripples: three rings, a third of a cycle apart, spread and fade.
      Repeater {
        model: 3
        Rectangle {
          id: ring
          required property int index
          property real size: 24
          width: size
          height: size
          x: -size / 2
          y: -size / 2
          radius: size / 2
          color: "transparent"
          border.color: root.green
          border.width: 2
          opacity: 0
          SequentialAnimation {
            running: root.shown
            loops: Animation.Infinite
            PauseAnimation { duration: ring.index * 600 }
            ParallelAnimation {
              NumberAnimation { target: ring; property: "size"; from: 24; to: 220; duration: 1800; easing.type: Easing.OutQuad }
              NumberAnimation { target: ring; property: "opacity"; from: 0.85; to: 0; duration: 1800; easing.type: Easing.OutQuad }
            }
            PauseAnimation { duration: (2 - ring.index) * 600 }
          }
        }
      }

      // The core: a soft halo and a bright centre that breathe together.
      Item {
        id: core
        SequentialAnimation on scale {
          running: root.shown
          loops: Animation.Infinite
          NumberAnimation { from: 0.9; to: 1.2; duration: 900; easing.type: Easing.InOutSine }
          NumberAnimation { from: 1.2; to: 0.9; duration: 900; easing.type: Easing.InOutSine }
        }
        Rectangle { width: 64; height: 64; x: -32; y: -32; radius: 32; color: root.green; opacity: 0.10 }
        Rectangle { width: 38; height: 38; x: -19; y: -19; radius: 19; color: root.green; opacity: 0.22 }
        Rectangle { width: 16; height: 16; x: -8; y: -8; radius: 8; color: root.green }
        Rectangle { width: 6; height: 6; x: -3; y: -3; radius: 3; color: "#eafff0" }
      }
    }
  }
}
