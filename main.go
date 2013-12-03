package main

import "os"
import "fmt"
import "strings"
import "github.com/vaughan0/go-ini"

type person struct {
    name string
    age  int
}

type Symbols struct {
    lock string
    network string
    separator string
    separator_thin string
}

type PowerlineAppendArgs struct {
    content string
    fg string
    bg string
    separator string
    separator_fg string
}

var colors = make(map[string] string)

var symbols = map[string] Symbols {
    "compatible": Symbols{
        lock: "RO",
        network: "SSH",
        separator: "\u25B6",
        separator_thin: "\u276F",
    },
    "patched": Symbols{
        lock: "\uE0A2",
        network: "\uE0A2",
        separator: "\uE0B0",
        separator_thin: "\uE0B1",
    },
    "flat": Symbols{
        lock: "",
        network: "",
        separator: "",
        separator_thin: "",
    },
}

var color_templates = map[string] string {
    "bash": "\\[\\e%s\\]",
    "zsh": "%%{%s%%}",
    "bare": "%s",
}

type PowerlineArgs struct {
    colorize_hostname bool
    cwd_max_depth int
    cwd_only bool
    mode string
    prev_error int
    shell string
}

type Powerline struct {
    args PowerlineArgs
    cwd string
    color_template string
    reset string
    lock string
    network string
    separator string
    separator_thin string
    segments []PowerlineAppendArgs
}

func (p *Powerline) SetColorTemplate() {
    //p.segments = make([20]PowerlineAppendArgs)
    //p.segments = make([]PowerlineAppendArgs)
    p.color_template = color_templates[p.args.shell]
}

func (p *Powerline) SetReset() {
    p.reset = fmt.Sprintf(p.color_template, "[0m")
}

func (p *Powerline) SetLock() {
    p.lock = symbols[p.args.mode].lock
}

func (p *Powerline) SetNetwork() {
    p.network = symbols[p.args.mode].network
}

func (p *Powerline) SetSeparator() {
    p.separator = symbols[p.args.mode].separator
}

func (p *Powerline) SetSeparatorThin() {
    p.separator_thin = symbols[p.args.mode].separator_thin
}

func (p Powerline) Color(prefix string, code string) string {
    return fmt.Sprintf(p.color_template, fmt.Sprintf("[%s;5;%sm", prefix, code))
}

func (p Powerline) FGColor(code string) string {
    return p.Color("38", code)
}

func (p Powerline) BGColor(code string) string {
    return p.Color("48", code)
}

func (p *Powerline) Append(args PowerlineAppendArgs) {
    if &args.separator == nil {
        args.separator = p.separator
    }

    if &args.separator_fg == nil {
        args.separator_fg = args.bg
    }

    p.segments = append(p.segments, args)

//        self.segments.append((content, fg, bg, separator or self.separator,
//            separator_fg or bg))
}

func (p Powerline) Draw() string {
    var out string = ""
    //return (''.join(self.draw_segment(i) for i in range(len(self.segments)))
    //            + self.reset).encode('utf-8')
    fmt.Println(p.segments)
    for idx, _ := range p.segments {
        out += p.DrawSegment(idx) + p.reset
    }
    return out
}

func (p Powerline) DrawSegment(idx int) string {
    var out string = ""
    var segment PowerlineAppendArgs = p.segments[idx]
    var next_segment PowerlineAppendArgs
    if (idx < len(p.segments) - 1) {
        next_segment = p.segments[idx + 1]
    }

    out += p.FGColor(segment.fg)
    out += p.BGColor(segment.bg)
    out += segment.content
    if &next_segment != nil {
        out += p.BGColor(next_segment.bg)
    } else {
        out += p.reset
    }
    out += p.FGColor(segment.separator_fg)
    out += segment.separator
//     segment = p.segments[idx]
//     next_segment = self.segments[idx + 1] if idx < len(self.segments)-1 else None
// 
//         return ''.join((
//             self.fgcolor(segment[1]),
//             self.bgcolor(segment[2]),
//             segment[0],
//             self.bgcolor(next_segment[2]) if next_segment else self.reset,
//             self.fgcolor(segment[4]),
//             segment[3]))
    return out
}

func get_valid_cwd() string {
    //    We check if the current working directory is valid or not. Typically
    //    happens when you checkout a different branch on git that doesn't have
    //    this directory.
    //    We return the original cwd because the shell still considers that to be
    //    the working directory, so returning our guess will confuse people
    //
    wd, err := os.Getwd()
    //fmt.Println("err:", err)
    //fmt.Println("wd:", wd)
    if err != nil {
        panic(err)
    }

    // TODO: write error handling here...
//     try:
//         cwd = os.getcwd()
//     except:
//         cwd = os.getenv('PWD')  # This is where the OS thinks we are
//         parts = cwd.split(os.sep)
//         up = cwd
//         while parts and not os.path.exists(up):
//             parts.pop()
//             up = os.sep.join(parts)
//         try:
//             os.chdir(up)
//         except:
//             warn("Your current directory is invalid.")
//             sys.exit(1)
//         warn("Your current directory is invalid. Lowest valid directory: " + up)
//     return cwd
    return wd
}

func (p *Powerline) AddUsernameSegment() {
    var user_prompt string = ""

    if (p.args.shell == "bash") {
        user_prompt = " \\u "
    } else if (p.args.shell == "zsh") {
        user_prompt = " %n "
    } else {
        user_prompt = fmt.Sprintf(" %s ", os.Getenv("USER"))
    }

    p.Append(PowerlineAppendArgs{content: user_prompt, fg: colors["USERNAME_FG"], bg: colors["USERNAME_BG"]})
}

func (p *Powerline) AddRootIndicatorSegment() {
    var root_indicators = map[string] string {
        "bash": " \\$ ",
        "zsh": " \\$ ",
        "bare": " $ ",
    }
    var bg = colors["CMD_PASSED_BG"]
    var fg = colors["CMD_PASSED_FG"]
    if p.args.prev_error != 0 {
        fg = colors["CMD_FAILED_FG"]
        bg = colors["CMD_FAILED_BG"]
    }

    p.Append(PowerlineAppendArgs{content: root_indicators[p.args.shell], fg: fg, bg: bg})
}


func GetShortPath(cwd string) []string {
    var home string = os.Getenv("HOME")
    fmt.Println(string(os.PathSeparator))
    fmt.Println(home)
    var names = strings.Split(cwd, string(os.PathSeparator))

    if names[0] == "" {
        names = names[1:]
    }

    var path string = ""
    for index := 1; index < len(names); index++ {
        path += string(os.PathSeparator) + names[index]
        path_dir, err := os.Stat(path)
        if err != nil {
            panic(err)
        }
        home_dir, err := os.Stat(home)
        if err != nil {
            panic(err)
        }
        if os.SameFile(path_dir, home_dir) {
            var ind int = index + 1
            fmt.Println(names[ind:])
            //return append([]string{"~"}, names[ind:])
        }
    }

    if len(names) == 0 {
        return []string{"~"}
    }
    return names
//     if names[0] == '': names = names[1:]
//     path = ''
//     for i in range(len(names)):
//         path += os.sep + names[i]
//         if os.path.samefile(path, home):
//             return ['~'] + names[i+1:]
//     if not names[0]:
//         return ['/']
//     return names
// 
// def add_cwd_segment():
//     cwd = powerline.cwd or os.getenv('PWD')
//     names = get_short_path(cwd.decode('utf-8'))
// 
//     max_depth = powerline.args.cwd_max_depth
//     if len(names) > max_depth:
//         names = names[:2] + [u'\u2026'] + names[2 - max_depth:]
// 
//     if not powerline.args.cwd_only:
//         for n in names[:-1]:
//             if n == '~' and Color.HOME_SPECIAL_DISPLAY:
//                 powerline.append(' %s ' % n, Color.HOME_FG, Color.HOME_BG)
//             else:
//                 powerline.append(' %s ' % n, Color.PATH_FG, Color.PATH_BG,
//                     powerline.separator_thin, Color.SEPARATOR_FG)
// 
//     if names[-1] == '~' and Color.HOME_SPECIAL_DISPLAY:
//         powerline.append(' %s ' % names[-1], Color.HOME_FG, Color.HOME_BG)
//     else:
//         powerline.append(' %s ' % names[-1], Color.CWD_FG, Color.PATH_BG)
}


func main() {
    configfile, err := ini.LoadFile("config")
    if err != nil {
        panic(err)
    }

    _, ok := configfile.Get("", "SEGMENTS")
    if !ok {
        panic("'SEGMENTS' variable missing from config")
    }

    theme, ok := configfile.Get("", "THEME")
    if !ok {
        panic("'THEME' variable missing from config")
    }

    // Load theme
    themefile, err := ini.LoadFile("themes/" + theme)
    if err != nil {
        panic(err)
    }

    for key, value := range themefile["COLORS"] {
        //fmt.Printf("%s => %s\n", key, value)
        colors[key] = value
    }

    GetShortPath(get_valid_cwd())

    // colorize_hostname=False, cwd_max_depth=5, cwd_only=False, mode='patched', prev_error=0, shell='bash'
    args := PowerlineArgs{colorize_hostname: false, cwd_max_depth: 5, cwd_only: false, mode: "patched", prev_error: 0, shell: "bash"}

    p := Powerline{args: args, cwd: get_valid_cwd()}
    p.SetColorTemplate()
    p.SetReset()
    p.SetLock()
    p.SetNetwork()
    p.SetSeparator()
    p.SetSeparatorThin()

    p.AddUsernameSegment()
    p.AddRootIndicatorSegment()
    fmt.Println(p)
    fmt.Println(p.Draw())

    
}
