import QtQuick
import QtQuick.Layouts
import Quickshell
import Quickshell.Hyprland
import Quickshell.Io
import qs.Commons
import qs.Ui

BarWidget {
  id: root
  moduleName: "omarchy.workspaces"

  // Onsen deviation: this bar shows its own monitor's workspaces, numbered by
  // their position on it, the way Super + number reaches them (Hyprland's r~N
  // selector), not every workspace by its global id. The window a pooled
  // widget sits in tells it the screen, and so the monitor.
  readonly property var panelWindow: root.QsWindow.window
  readonly property var monitor: panelWindow && panelWindow.screen ? Hyprland.monitorFor(panelWindow.screen) : null

  function monitorWorkspaces() {
    var out = []
    var values = Hyprland.workspaces.values
    var name = root.monitor ? root.monitor.name : ""

    for (var i = 0; i < values.length; i++) {
      var workspace = values[i]
      if (workspace.id <= 0) continue
      if (!workspace.monitor || workspace.monitor.name !== name) continue
      out.push(workspace)
    }

    out.sort(function(left, right) { return left.id - right.id })
    return out
  }

  // Positions 1..N on this monitor, at least five, so the row keeps a size.
  function positions() {
    var count = root.monitorWorkspaces().length
    var out = []
    for (var position = 1; position <= Math.max(5, count); position++) out.push(position)
    return out
  }

  function workspaceAt(position) {
    var workspaces = root.monitorWorkspaces()
    return position >= 1 && position <= workspaces.length ? workspaces[position - 1] : null
  }

  function positionOf(workspace) {
    var workspaces = root.monitorWorkspaces()
    for (var i = 0; i < workspaces.length; i++) {
      if (workspaces[i].id === workspace.id) return i + 1
    }
    return -1
  }

  // The monitor's own active workspace, which may not hold the pointer's focus.
  function activePosition() {
    var active = root.monitor ? root.monitor.activeWorkspace : null
    return active ? root.positionOf(active) : -1
  }

  function focusPosition(position) {
    if (!root.bar) return
    var workspace = root.workspaceAt(position)
    if (workspace) {
      root.bar.run("hyprctl dispatch " + Util.shellQuote("hl.dsp.focus({ workspace = \"" + workspace.id + "\" })"))
      return
    }
    // A position past the last workspace is a new one: r~N counts on the focused
    // monitor, so focus this bar's monitor first.
    if (!root.monitor) return
    root.bar.run("hyprctl --batch " + Util.shellQuote(
      "dispatch hl.dsp.focus({ monitor = \"" + root.monitor.name + "\" }) ; " +
      "dispatch hl.dsp.focus({ workspace = \"r~" + position + "\" })"))
  }

  // A wheel notch moves one entry along the row, and a touchpad adds up its
  // smaller deltas until they make a notch's worth.
  property real wheelAccumulator: 0

  function scrollWorkspaces(delta) {
    var wheel = Util.wheelSteps(root.wheelAccumulator, delta)
    root.wheelAccumulator = wheel.remainder
    if (wheel.steps === 0) return

    // Holds at either end rather than wrapping, so a touchpad swipe that runs
    // long stops on the last workspace.
    var positions = root.positions()
    var index = positions.indexOf(root.activePosition())
    if (index === -1) return
    var target = positions[Util.clamp(index - wheel.steps, 0, positions.length - 1)]
    if (target !== positions[index]) root.focusPosition(target)
  }

  // The focused workspace shows its layout: a square tiles, a circle scrolls,
  // and a triangle floats. Clicking it cycles to the next one, as Super + L does.
  property string focusedMode: "dwindle"
  property bool modeStale: false

  function modeGlyph(mode) {
    if (mode === "scrolling") return "\uDB81\uDF65"
    if (mode === "floating") return "\uDB81\uDD36"
    return "\uDB85\uDCFB"
  }

  // A layout only changes through the layout commands, which ask for a refresh,
  // or a config reload, so each workspace's mode is read once and cached rather
  // than running the script on every workspace switch.
  property var modeCache: ({})
  property int modeGeneration: 0
  property int modeProcWorkspace: -1
  property int modeProcGeneration: 0

  function focusedWorkspaceId() {
    return Hyprland.focusedWorkspace ? Hyprland.focusedWorkspace.id : -1
  }

  function showFocusedMode() {
    var mode = root.modeCache[root.focusedWorkspaceId()]
    if (mode !== undefined) root.focusedMode = mode
    else readMode()
  }

  function readMode() {
    if (modeProc.running) {
      root.modeStale = true
      return
    }
    var id = root.focusedWorkspaceId()
    if (id === -1) return
    root.modeProcWorkspace = id
    root.modeProcGeneration = root.modeGeneration
    modeProc.command = ["omarchy-hyprland-workspace-layout-current", String(id)]
    modeProc.running = true
  }

  function refresh() {
    root.modeCache = ({})
    root.modeGeneration++
    readMode()
  }

  function cycleLayout() {
    if (root.bar) root.bar.run("omarchy-hyprland-workspace-layout-toggle")
  }

  Component.onCompleted: refresh()

  Connections {
    target: Hyprland
    function onFocusedWorkspaceChanged() { root.showFocusedMode() }
    function onRawEvent(event) {
      if (event && String(event.name) === "configreloaded") root.refresh()
    }
  }

  Process {
    id: modeProc
    stdout: StdioCollector {
      onStreamFinished: {
        var mode = text.trim()
        // A read that started before a refresh may predate the change.
        if (!mode || root.modeProcGeneration !== root.modeGeneration) return
        root.modeCache[root.modeProcWorkspace] = mode
        if (root.modeProcWorkspace === root.focusedWorkspaceId()) root.focusedMode = mode
      }
    }
    onExited: {
      if (!root.modeStale) return
      root.modeStale = false
      // Deferred: the process still reports running while it signals exit.
      Qt.callLater(root.showFocusedMode)
    }
  }

  // Super + L and the default layout change the mode without a focus change,
  // so they ask every bar to read it again.
  ShellIpc {
    target: "omarchy.workspaces"

    function refresh(): void { root.broadcast("refresh") }
  }

  readonly property real trailingGap: root.vertical ? 0 : Style.spaceReal(1.5)

  implicitWidth: grid.implicitWidth + trailingGap
  implicitHeight: grid.implicitHeight

  GridLayout {
    id: grid
    anchors.fill: parent
    anchors.rightMargin: root.trailingGap
    columns: root.vertical ? 1 : root.positions().length
    columnSpacing: root.vertical ? 0 : Style.space(1)
    rowSpacing: root.vertical ? Style.space(2) : 0

    Repeater {
      model: root.positions()

      WidgetButton {
        required property int modelData

        readonly property var workspace: root.workspaceAt(modelData)
        readonly property bool occupied: workspace !== null && workspace.toplevels.values.length > 0
        readonly property bool focused: workspace !== null && Hyprland.focusedWorkspace !== null && Hyprland.focusedWorkspace.id === workspace.id
        // Not `active`: WidgetButton's own `active` paints the urgent colour.
        readonly property bool shownOnMonitor: modelData === root.activePosition()

        bar: root.bar
        text: focused ? root.modeGlyph(root.focusedMode) : (modelData === 10 ? "0" : String(modelData))
        opacity: occupied || focused || shownOnMonitor ? 1 : 0.5
        horizontalMargin: 6
        verticalPadding: 6
        fixedWidth: root.vertical ? root.barSize : Style.space(20)
        fixedHeight: root.barSize
        onPressed: function() {
          if (focused) root.cycleLayout()
          else root.focusPosition(modelData)
        }
        onWheelMoved: function(delta) { root.scrollWorkspaces(delta) }
      }
    }
  }
}
