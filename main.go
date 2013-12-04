package main

import "os"
import "fmt"
import "path"
import "path/filepath"
import "bytes"
import "strings"
import "strconv"
import "os/exec"
import "github.com/vaughan0/go-ini"

type GitStatus struct {
    has_pending_commits bool
    has_untracked_files bool
    origin_position string
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
    out := ""

    for idx, _ := range p.segments {
        out += p.DrawSegment(idx)
    }
    out += p.reset
    return out
}

func (p Powerline) DrawSegment(idx int) string {
    out := ""
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
    env := os.Getenv("VIRTUAL_ENV")
    if env == "" {
        return
    }

    env_name := path.Base(env)
    content := fmt.Sprintf(" %s ", env_name)

    p.Append(PowerlineAppendArgs{content: content, fg: colors["VIRTUAL_ENV_FG"], bg: colors["VIRTUAL_ENV_BG"]})
}

func (p *Powerline) AddUsernameSegment() {
    user_prompt := ""

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
    host_prompt := ""
    hostname, err := os.Hostname()
    if err != nil {
        panic(err)
    }

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
    ssh_client := os.Getenv("SSH_CLIENT")
    content := fmt.Sprintf(" %s ", p.network)
    if ssh_client != "" {
        p.Append(PowerlineAppendArgs{content: content, fg: colors["SSH_FG"], bg: colors["SSH_BG"]})
    }
}

func GetShortPath(cwd string) []string {
    home := os.Getenv("HOME")
    home_dir, err := os.Stat(home)
    if err != nil {
        panic(err)
    }

    names := strings.Split(cwd, string(os.PathSeparator))

    if names[0] == "" {
        names = names[1:]
    }

    path := ""
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
    if p.cwd != "" {
        cwd = p.cwd
    } else {
        cwd = os.Getenv("PWD")
    }

    names = GetShortPath(cwd)
    max_depth := p.args.cwd_max_depth

    if len(names) > max_depth {
        //names = names[:2] + ["\u2026"] + names[2 - max_depth:]
    }

    home_special_display, err := strconv.ParseBool(colors["HOME_SPECIAL_DISPLAY"])
    if err != nil {
        panic(err)
    }

    if p.args.cwd_only != true {
        for _, n := range names[:len(names)-1] {
            content := fmt.Sprintf(" %s ", n)
            if n == "~" && home_special_display {
                p.Append(PowerlineAppendArgs{content: content, fg: colors["HOME_FG"], bg: colors["HOME_BG"]})
            } else {
                p.Append(PowerlineAppendArgs{content: content, fg: colors["PATH_FG"], bg: colors["PATH_BG"],
                    separator: p.separator_thin, separator_fg: colors["SEPARATOR_FG"]})
            }
        }
    }

    content := fmt.Sprintf(" %s ", names[len(names)-1])

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


func GetGitStatus() GitStatus {
    var gitstatus GitStatus
    gitstatus.has_pending_commits = true
    gitstatus.has_untracked_files = false

    cmd := exec.Command("git", "status", "--ignore-submodules")

    var out bytes.Buffer
    cmd.Stdout = &out
    err := cmd.Run()
    if err != nil {
        panic(err)
    }

    for _, line := range strings.Split(out.String(), "\n") {
        occurence := strings.Index(line, "Your branch is ")
        if occurence != -1 {
            parts := strings.Split(line[occurence + len("Your branch is "):], " ")

            by_occurence := strings.Index(line[occurence + len("Your branch is "):], " by ")
            commit_occurence := strings.Index(line[occurence + len("Your branch is "):], " commit")

            commits_cnt := line[occurence + len("Your branch is "):][by_occurence + len(" by "):commit_occurence]
            gitstatus.origin_position += fmt.Sprintf(" %s", strings.TrimSpace(commits_cnt))

            if parts[0] == "behind" {
                gitstatus.origin_position += "\u21E3"
            } else if parts[0] == "ahead" {
                gitstatus.origin_position += "\u21E1"
            }
        }

        if(strings.Contains(line, "nothing to commit")) {
            gitstatus.has_pending_commits = false
        }
        if(strings.Contains(line, "Untracked files")) {
            gitstatus.has_untracked_files = true
        }        
    }

    return gitstatus
}

func (p *Powerline) AddGitSegment() {
    // git branch 2> /dev/null | grep -e '\\*'
    cmd_git_branch := exec.Command("git", "branch", "--no-color")
    var git_branch_out bytes.Buffer
    cmd_git_branch.Stdout = &git_branch_out
    err := cmd_git_branch.Run()
    if err != nil {
        //panic(err)
        return
    }

    cmd_grep := exec.Command("grep", "-e", "\\*")
    cmd_grep.Stdin = &git_branch_out

    var grep_out bytes.Buffer
    cmd_grep.Stdout = &grep_out
    err2 := cmd_grep.Run()
    if err2 != nil {
        panic(err2)
    }

    if grep_out.String() == "" {
        return
    }

    branch := strings.TrimSpace(grep_out.String()[2:])
    gitstatus := GetGitStatus()
    branch += gitstatus.origin_position
    if gitstatus.has_untracked_files {
        branch += " +"
    }

    bg := colors["REPO_CLEAN_BG"]
    fg := colors["REPO_CLEAN_FG"]
    if gitstatus.has_pending_commits {
        bg = colors["REPO_DIRTY_BG"]
        fg = colors["REPO_DIRTY_FG"]
    }

    content := fmt.Sprintf(" %s ", branch)

    p.Append(PowerlineAppendArgs{content: content, fg: fg, bg: bg})
}

func (p *Powerline) AddRootIndicatorSegment() {
    root_indicators := map[string] string {
        "bash": " \\$ ",
        "zsh": " \\$ ",
        "bare": " $ ",
    }
    bg := colors["CMD_PASSED_BG"]
    fg := colors["CMD_PASSED_FG"]
    if p.args.prev_error != 0 {
        fg = colors["CMD_FAILED_FG"]
        bg = colors["CMD_FAILED_BG"]
    }

    p.Append(PowerlineAppendArgs{content: root_indicators[p.args.shell], fg: fg, bg: bg})
}


func main() {
    exec_path := os.Args[0]
    abs_exec_path, err := filepath.Abs(exec_path)
    if err != nil {
        panic(err)
    }

    configfile, err := ini.LoadFile(fmt.Sprintf("%s/config", filepath.Dir(abs_exec_path)))
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
    themefile, err := ini.LoadFile(fmt.Sprintf("%s/themes/%s", filepath.Dir(abs_exec_path), theme))
    if err != nil {
        panic(err)
    }

    for key, value := range themefile["COLORS"] {
        colors[key] = value
    }

    var prev_error int = 0
    if len(os.Args) > 1 {
        prev_error, err = strconv.Atoi(os.Args[1])
        if err != nil {
            panic(err)
        }
    }
    
    // colorize_hostname=False, cwd_max_depth=5, cwd_only=False, mode='patched', prev_error=0, shell='bash'
    args := PowerlineArgs{colorize_hostname: false, cwd_max_depth: 5, cwd_only: false, mode: "patched", prev_error: prev_error, shell: "bash"}

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
    p.AddGitSegment()
    p.AddRootIndicatorSegment()

    fmt.Printf(p.Draw())
}
