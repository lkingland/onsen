#!/usr/bin/env python3
"""Where the focus beacon goes: one JSON line, {"monitor", "x", "y"}.

x and y are logical pixels on the focused monitor. The point is the centre of
the focused window, or of the screen when it has no window. When the focused
window is a terminal holding a tmux client, the point is the centre of that
client's active pane instead: tmux reports the pane in character cells, and
the window's size over the client's size in cells gives a cell.
"""
import json
import subprocess


def run(*args):
    try:
        return subprocess.run(args, capture_output=True, text=True, timeout=2).stdout
    except Exception:
        return ""


def ppid(pid):
    try:
        with open(f"/proc/{pid}/stat") as f:
            return int(f.read().rsplit(")", 1)[1].split()[1])
    except Exception:
        return 0


def descends_from(pid, ancestor):
    for _ in range(64):
        if pid == ancestor:
            return True
        if pid <= 1:
            return False
        pid = ppid(pid)
    return False


def tmux_pane_centre(window_pid):
    """Fractions (fx, fy) of the window for the active pane's centre, or None."""
    for line in run("tmux", "list-clients", "-F", "#{client_pid} #{client_tty}").splitlines():
        try:
            pid, tty = line.split(" ", 1)
        except ValueError:
            continue
        if not descends_from(int(pid), window_pid):
            continue
        fields = run("tmux", "display", "-c", tty, "-p",
                     "#{pane_left} #{pane_top} #{pane_width} #{pane_height} "
                     "#{client_width} #{client_height} #{window_height} #{status-position}").split()
        if len(fields) != 8:
            return None
        left, top, width, height, cw, ch, wh = (int(v) for v in fields[:7])
        if cw <= 0 or ch <= 0:
            return None
        # The status lines sit above the panes when the status is at the top.
        offset = ch - wh if fields[7] == "top" else 0
        return ((left + width / 2) / cw, (offset + top + height / 2) / ch)
    return None


def main():
    try:
        monitors = json.loads(run("hyprctl", "-j", "monitors") or "[]")
        window = json.loads(run("hyprctl", "-j", "activewindow") or "{}")
    except ValueError:
        return
    mon = next((m for m in monitors if m.get("focused")), None)
    if not mon:
        return
    scale = mon.get("scale") or 1
    x, y = mon["width"] / scale / 2, mon["height"] / scale / 2

    if window.get("at") and window.get("size") and window.get("monitor") == mon.get("id"):
        (wx, wy), (ww, wh) = window["at"], window["size"]
        fx, fy = 0.5, 0.5
        if window.get("pid"):
            pane = tmux_pane_centre(int(window["pid"]))
            if pane:
                fx, fy = pane
        x, y = wx + ww * fx - mon["x"], wy + wh * fy - mon["y"]

    print(json.dumps({"monitor": mon["name"], "x": round(x, 1), "y": round(y, 1)}))


if __name__ == "__main__":
    main()
