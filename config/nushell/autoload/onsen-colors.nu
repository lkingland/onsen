# Onsen palette for nushell: table and syntax colors, the vi prompt marks,
# ls/eza colors, bat's theme, and eza's theme directory. Nushell autoloads
# this after config.nu (its user autoload directory).

$env.config.color_config = {
    separator: "#606060"
    leading_trailing_space_bg: { attr: n }
    header: { fg: "#E0E0E0", attr: b }
    row_index: "#606060"
    record: "#E0E0E0"
    list: "#E0E0E0"
    bool: "#8FD9A8"
    int: "#B0B0B0"
    float: "#B0B0B0"
    filesize: "#B0B0B0"
    duration: "#B0B0B0"
    date: "#B0B0B0"
    datetime: "#B0B0B0"
    range: "#B0B0B0"
    string: "#E0E0E0"
    nothing: "#606060"
    binary: "#B0B0B0"
    cellpath: "#A8C4F7"
    hints: "#606060"
    shape_binary: "#B0B0B0"
    shape_block: "#A8C4F7"
    shape_bool: "#8FD9A8"
    shape_datetime: "#B0B0B0"
    shape_directory: "#A8C4F7"
    shape_external: "#8FD9A8"
    shape_filepath: "#E0E0E0"
    shape_flag: "#A8C4F7"
    shape_float: "#B0B0B0"
    shape_int: "#B0B0B0"
    shape_list: "#E0E0E0"
    shape_literal: "#8FD9A8"
    shape_nothing: "#606060"
    shape_operator: "#E0E0E0"
    shape_range: "#B0B0B0"
    shape_record: "#E0E0E0"
    shape_string: "#E0E0E0"
    shape_variable: "#8FD9A8"
}

let green_bold = { fg: "#8FD9A8", attr: b }
let blue_bold = { fg: "#A8C4F7", attr: b }
$env.PROMPT_INDICATOR_VI_INSERT = $"(ansi $green_bold)❯(ansi reset) "
$env.PROMPT_INDICATOR_VI_NORMAL = $"(ansi $blue_bold)❮(ansi reset) "

$env.LS_COLORS = ([
    "di=1;38;2;168;196;247"       # directories: bold blue
    "ln=38;2;168;196;247"         # symlinks: blue
    "ex=38;2;255;158;80"          # executables: orange
    "fi=38;2;224;224;224"         # regular files: white
    "pi=38;2;96;96;96"            # pipes: grey
    "so=38;2;96;96;96"            # sockets: grey
    "bd=38;2;255;158;80"          # block devices: orange
    "cd=38;2;255;158;80"          # char devices: orange
    "or=38;2;255;107;107"         # orphan (broken symlink): red
    "mi=38;2;255;107;107"         # missing file: red
    "su=38;2;255;107;107"         # setuid: red
    "sg=38;2;255;158;80"          # setgid: orange
    "tw=1;38;2;168;196;247"       # sticky other-writable: bold blue
    "ow=38;2;168;196;247"         # other-writable: blue
    # Binary files: green
    "*.png=38;2;143;217;168"
    "*.jpg=38;2;143;217;168"
    "*.jpeg=38;2;143;217;168"
    "*.gif=38;2;143;217;168"
    "*.webp=38;2;143;217;168"
    "*.svg=38;2;143;217;168"
    "*.ico=38;2;143;217;168"
    "*.mp4=38;2;143;217;168"
    "*.mkv=38;2;143;217;168"
    "*.avi=38;2;143;217;168"
    "*.mov=38;2;143;217;168"
    "*.webm=38;2;143;217;168"
    "*.mp3=38;2;143;217;168"
    "*.flac=38;2;143;217;168"
    "*.wav=38;2;143;217;168"
    "*.ogg=38;2;143;217;168"
    "*.zip=38;2;143;217;168"
    "*.tar=38;2;143;217;168"
    "*.gz=38;2;143;217;168"
    "*.xz=38;2;143;217;168"
    "*.7z=38;2;143;217;168"
    "*.rar=38;2;143;217;168"
    "*.pdf=38;2;143;217;168"
    "*.o=38;2;143;217;168"
    "*.so=38;2;143;217;168"
    "*.a=38;2;143;217;168"
    # Temp files: dark grey
    "*.tmp=38;2;96;96;96"
    "*.swp=38;2;96;96;96"
    "*.bak=38;2;96;96;96"
] | str join ":")
$env.BAT_THEME = "onsen"
$env.EZA_CONFIG_DIR = $"($env.HOME)/.config/omarchy/current/theme"
