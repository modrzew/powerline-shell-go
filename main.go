package main

import "os"
import "fmt"
import "path"
import "strings"
import "strconv"
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
    if args.separator == "" {
        args.separator = p.separator
    }

    if args.separator_fg == "" {
        args.separator_fg = args.bg
    }

    p.segments = append(p.segments, args)
}

func (p Powerline) Draw() string {
    var out string = ""
    //return (''.join(self.draw_segment(i) for i in range(len(self.segments)))
    //            + self.reset).encode('utf-8')

    for idx, _ := range p.segments {
        out += p.DrawSegment(idx)
    }
    out += p.reset
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
    if next_segment.content != "" {
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

func GetValidCwd() string {
    //    We check if the current working directory is valid or not. Typically
    //    happens when you checkout a different branch on git that doesn't have
    //    this directory.
    //    We return the original cwd because the shell still considers that to be
    //    the working directory, so returning our guess will confuse people
    //
    wd, err := os.Getwd()
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

func (p *Powerline) AddVirtualEnvSegment() {
    var env string = os.Getenv("VIRTUAL_ENV")
    if env == "" {
        return
    }

    var env_name string = path.Base(env)
    var content string = fmt.Sprintf(" %s ", env_name)

    p.Append(PowerlineAppendArgs{content: content, fg: colors["VIRTUAL_ENV_FG"], bg: colors["VIRTUAL_ENV_BG"]})
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

func (p *Powerline) AddHostnameSegment() {
    var host_prompt string
    hostname, err := os.Hostname()
    if err != nil {
        panic(err)
    }
    //fmt.Println("hostname")
    //fmt.Println(hostname)

    if p.args.colorize_hostname {

        //from lib.color_compliment import stringToHashToColorAndOpposite
        //from lib.colortrans import rgb2short
        //FG, BG = stringToHashToColorAndOpposite(hostname)
        //FG, BG = (rgb2short(*color) for color in [FG, BG])
        //host_prompt = ' %s' % hostname.split('.')[0]

        //powerline.append(host_prompt, FG, BG)
    } else {
        if p.args.shell == "bash" {
            host_prompt = " \\h "
        } else if p.args.shell == "zsh" {
            host_prompt = " %m "
        } else {
            host_prompt = fmt.Sprintf(" %s ", strings.Split(hostname, ".")[0])
        }

        p.Append(PowerlineAppendArgs{content: host_prompt, fg: colors["HOSTNAME_FG"], bg: colors["HOSTNAME_BG"]})
    }
}

func (p *Powerline) AddSshSegment() {
    var ssh_client string = os.Getenv("SSH_CLIENT")
    var content string = fmt.Sprintf(" %s ", p.network)
    if ssh_client != "" {
        p.Append(PowerlineAppendArgs{content: content, fg: colors["SSH_FG"], bg: colors["SSH_BG"]})
    }
}

func GetShortPath(cwd string) []string {
    var home string = os.Getenv("HOME")
    home_dir, err := os.Stat(home)
    if err != nil {
        panic(err)
    }

    var names = strings.Split(cwd, string(os.PathSeparator))

    if names[0] == "" {
        names = names[1:]
    }

    var path string = ""
    for index := 0; index < len(names); index++ {
        path += string(os.PathSeparator) + names[index]

        path_dir, err := os.Stat(path)
        if err != nil {
            panic(err)
        }

        if os.SameFile(path_dir, home_dir) {
            var ind int = index + 1
            return append([]string{"~"}, names[ind:]...)
        }
    }

    if len(names) == 0 {
        return []string{"~"}
    }
    return names
}

func (p *Powerline) AddCwdSegment() {
    var cwd string
    var names []string
    var max_depth int
    if p.cwd != "" {
        cwd = p.cwd
    } else {
        cwd = os.Getenv("PWD")
    }

    names = GetShortPath(cwd)
    max_depth = p.args.cwd_max_depth

    if len(names) > max_depth {
        //names = names[:2] + ["\u2026"] + names[2 - max_depth:]
    }

    home_special_display, err := strconv.ParseBool(colors["HOME_SPECIAL_DISPLAY"])
    if err != nil {
        panic(err)
    }

    if p.args.cwd_only != true {
        for _, n := range names[:len(names)-1] {
            var content string = fmt.Sprintf(" %s ", n)
            if n == "~" && home_special_display {
                p.Append(PowerlineAppendArgs{content: content, fg: colors["HOME_FG"], bg: colors["HOME_BG"]})
            } else {
                p.Append(PowerlineAppendArgs{content: content, fg: colors["PATH_FG"], bg: colors["PATH_BG"],
                    separator: p.separator_thin, separator_fg: colors["SEPARATOR_FG"]})
            }
        }
    }

    var content string = fmt.Sprintf(" %s ", names[len(names)-1])

    if names[len(names)-1] == "~" && home_special_display {
        p.Append(PowerlineAppendArgs{content: content, fg: colors["HOME_FG"], bg: colors["HOME_BG"]})
    } else {
        p.Append(PowerlineAppendArgs{content: content, fg: colors["CWD_FG"], bg: colors["PATH_BG"]})
    }
}

func (p *Powerline) AddReadOnlySegment() {
//     var cwd string
//     if p.cwd != "" {
//         cwd = p.cwd
//     } else {
//         cwd = os.Getenv("PWD")
//     }

    
    // TODO: fix that
    //if not os.access(cwd, os.W_OK):
    //    powerline.append(' %s ' % powerline.lock, Color.READONLY_FG, Color.READONLY_BG)
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

    GetShortPath(GetValidCwd())

    // colorize_hostname=False, cwd_max_depth=5, cwd_only=False, mode='patched', prev_error=0, shell='bash'
    args := PowerlineArgs{colorize_hostname: false, cwd_max_depth: 5, cwd_only: false, mode: "patched", prev_error: 0, shell: "bash"}

    p := Powerline{args: args, cwd: GetValidCwd()}
    p.SetColorTemplate()
    p.SetReset()
    p.SetLock()
    p.SetNetwork()
    p.SetSeparator()
    p.SetSeparatorThin()

    p.AddVirtualEnvSegment()
    p.AddUsernameSegment()
    p.AddHostnameSegment()
    p.AddSshSegment()
    p.AddCwdSegment()
    p.AddRootIndicatorSegment()
    //fmt.Println(p)
    fmt.Println(p.Draw())

    
}
