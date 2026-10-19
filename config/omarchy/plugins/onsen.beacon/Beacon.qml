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
// it follows focus to wherever a chord moves it. Visual only: no keyboard
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

  // hyprctl reports window and monitor positions in the same layout
  // coordinates, so the window's centre less its monitor's origin is a point
  // on the overlay, which covers that monitor.
  function place(text) {
    var parts = String(text || "").split("\n--onsen-beacon--\n")
    var win = {}, monitors = []
    try { win = JSON.parse(parts[0] || "{}") } catch (e) {}
    try { monitors = JSON.parse(parts[1] || "[]") } catch (e) {}

    var mon = null
    for (var i = 0; i < monitors.length; i++) if (monitors[i].focused) mon = monitors[i]
    if (!mon) return
    var scale = mon.scale > 0 ? mon.scale : 1
    var width = mon.width / scale, height = mon.height / scale

    if (win && win.at && win.size && win.monitor === mon.id) {
      cx = win.at[0] + win.size[0] / 2 - mon.x
      cy = win.at[1] + win.size[1] / 2 - mon.y
    } else {
      cx = width / 2
      cy = height / 2
    }

    // Focus moved to another screen mid-hold: the overlay maps on the focused
    // screen when it opens, so close and reopen it there.
    if (monitorName !== "" && mon.name !== monitorName && shown) {
      shown = false
      Qt.callLater(function() { root.shown = root.held })
    }
    monitorName = mon.name
  }

  Process {
    id: locate
    command: ["sh", "-c", "hyprctl -j activewindow; printf '\\n--onsen-beacon--\\n'; hyprctl -j monitors"]
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
