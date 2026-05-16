package client

import (
	"context"
	"encoding/json"
)

// daemonResult wraps the goposix daemon's embedded result envelope returned
// from utility methods (goposix.<name>).
type daemonResult struct {
	ExitCode int             `json:"exitCode"`
	Data     json.RawMessage `json:"data"`
}

func callUtility[T any](c *Client, ctx context.Context, method string, params interface{}) (*T, error) {
	var raw daemonResult
	if err := c.Call(ctx, method, params, &raw); err != nil {
		return nil, err
	}
	var data T
	if raw.Data != nil {
		if err := json.Unmarshal(raw.Data, &data); err != nil {
			return nil, err
		}
	}
	return &data, nil
}

// --- Result types and helpers ---

// BasenameResult is the output of basename --json.
type BasenameResult struct {
	Result string `json:"result"`
}

// Basename runs goposix.basename.
func (c *Client) Basename(ctx context.Context, path string) (*BasenameResult, error) {
	return callUtility[BasenameResult](c, ctx, "goposix.basename", map[string]string{"path": path})
}

// DirnameResult is the output of dirname --json.
type DirnameResult struct {
	Result string `json:"result"`
}

// Dirname runs goposix.dirname.
func (c *Client) Dirname(ctx context.Context, path string) (*DirnameResult, error) {
	return callUtility[DirnameResult](c, ctx, "goposix.dirname", map[string]string{"path": path})
}

// EchoResult is the output of echo --json.
type EchoResult struct {
	Text string `json:"text"`
}

// Echo runs goposix.echo.
func (c *Client) Echo(ctx context.Context, text string) (*EchoResult, error) {
	return callUtility[EchoResult](c, ctx, "goposix.echo", map[string]string{"text": text})
}

// HostnameResult is the output of hostname --json.
type HostnameResult struct {
	Hostname string `json:"hostname"`
}

// Hostname runs goposix.hostname.
func (c *Client) Hostname(ctx context.Context) (*HostnameResult, error) {
	return callUtility[HostnameResult](c, ctx, "goposix.hostname", nil)
}

// PrintfResult is the output of printf --json.
type PrintfResult struct {
	Output string `json:"output"`
}

// Printf runs goposix.printf. format and args work like C printf.
func (c *Client) Printf(ctx context.Context, format string, args ...string) (*PrintfResult, error) {
	flags := append([]string{format}, args...)
	return callUtility[PrintfResult](c, ctx, "goposix.printf", map[string]interface{}{"flags": flags})
}

// PwdResult is the output of pwd --json.
type PwdResult struct {
	Path string `json:"path"`
}

// Pwd runs goposix.pwd.
func (c *Client) Pwd(ctx context.Context) (*PwdResult, error) {
	return callUtility[PwdResult](c, ctx, "goposix.pwd", nil)
}

// TestResult is the output of test --json.
type TestResult struct {
	Result bool `json:"result"`
}

// Test runs goposix.test with the given flags.
func (c *Client) Test(ctx context.Context, flags []string) (*TestResult, error) {
	return callUtility[TestResult](c, ctx, "goposix.test", map[string]interface{}{"flags": flags})
}

// WhoamiResult is the output of whoami --json.
type WhoamiResult struct {
	User string `json:"user"`
	UID  int    `json:"uid"`
}

// Whoami runs goposix.whoami.
func (c *Client) Whoami(ctx context.Context) (*WhoamiResult, error) {
	return callUtility[WhoamiResult](c, ctx, "goposix.whoami", nil)
}

// ReadlinkResult is the output of readlink --json.
type ReadlinkResult struct {
	Path   string `json:"path"`
	Target string `json:"target"`
}

// Readlink runs goposix.readlink.
func (c *Client) Readlink(ctx context.Context, path string) (*ReadlinkResult, error) {
	return callUtility[ReadlinkResult](c, ctx, "goposix.readlink", map[string]string{"path": path})
}

// IDInfo is the output of id --json.
type IDInfo struct {
	UID    int      `json:"uid"`
	User   string   `json:"user"`
	GID    int      `json:"gid"`
	Group  string   `json:"group"`
	Groups []string `json:"groups"`
}

// ID runs goposix.id.
func (c *Client) ID(ctx context.Context) (*IDInfo, error) {
	return callUtility[IDInfo](c, ctx, "goposix.id", nil)
}

// DateInfo is the output of date --json.
type DateInfo struct {
	ISO      string `json:"iso"`
	Unix     int64  `json:"unix"`
	UTC      string `json:"utc"`
	Timezone string `json:"timezone"`
}

// Date runs goposix.date.
func (c *Client) Date(ctx context.Context) (*DateInfo, error) {
	return callUtility[DateInfo](c, ctx, "goposix.date", nil)
}

// UnameResult is the output of uname --json.
type UnameResult struct {
	Sysname  string `json:"sysname"`
	Nodename string `json:"nodename"`
	Release  string `json:"release"`
	Version  string `json:"version"`
	Machine  string `json:"machine"`
}

// Uname runs goposix.uname.
func (c *Client) Uname(ctx context.Context) (*UnameResult, error) {
	return callUtility[UnameResult](c, ctx, "goposix.uname", nil)
}

// EnvVarsResult is the output of env/printenv --json.
type EnvVarsResult struct {
	Vars map[string]string `json:"vars"`
}

// Env runs goposix.env.
func (c *Client) Env(ctx context.Context, flags []string, vars map[string]string) (*EnvVarsResult, error) {
	return callUtility[EnvVarsResult](c, ctx, "goposix.env", map[string]interface{}{
		"flags": flags,
		"vars":  vars,
	})
}

// Printenv runs goposix.printenv.
func (c *Client) Printenv(ctx context.Context, name string) (*EnvVarsResult, error) {
	return callUtility[EnvVarsResult](c, ctx, "goposix.printenv", map[string]interface{}{"flags": []string{name}})
}

// CatResult is the output of cat --json.
type CatResult struct {
	Lines     []string `json:"lines"`
	LineCount int      `json:"lineCount"`
}

// Cat runs goposix.cat.
func (c *Client) Cat(ctx context.Context, path string) (*CatResult, error) {
	return callUtility[CatResult](c, ctx, "goposix.cat", map[string]string{"path": path})
}

// HeadResult is the output of head --json.
type HeadResult struct {
	Lines     []string `json:"lines"`
	LineCount int      `json:"lineCount"`
}

// Head runs goposix.head. If n is 0, defaults to 10.
func (c *Client) Head(ctx context.Context, path string, n int) (*HeadResult, error) {
	flags := []string{}
	if n > 0 {
		flags = append(flags, "-n", itoa(n))
	}
	return callUtility[HeadResult](c, ctx, "goposix.head", map[string]interface{}{"path": path, "flags": flags})
}

// TailResult is the output of tail --json.
type TailResult struct {
	Lines     []string `json:"lines"`
	LineCount int      `json:"lineCount"`
}

// Tail runs goposix.tail. If n is 0, defaults to 10.
func (c *Client) Tail(ctx context.Context, path string, n int) (*TailResult, error) {
	flags := []string{}
	if n > 0 {
		flags = append(flags, "-n", itoa(n))
	}
	return callUtility[TailResult](c, ctx, "goposix.tail", map[string]interface{}{"path": path, "flags": flags})
}

// SortResult is the output of sort --json.
type SortResult struct {
	Lines []string `json:"lines"`
	Count int      `json:"count"`
}

// Sort runs goposix.sort on the provided input lines.
func (c *Client) Sort(ctx context.Context, flags []string) (*SortResult, error) {
	return callUtility[SortResult](c, ctx, "goposix.sort", map[string]interface{}{"flags": flags})
}

// CutResult is the output of cut --json.
type CutResult struct {
	Lines []CutLine `json:"lines"`
}

// CutLine is a single line result from cut.
type CutLine struct {
	Fields []string `json:"fields"`
}

// Cut runs goposix.cut.
func (c *Client) Cut(ctx context.Context, flags []string) (*CutResult, error) {
	return callUtility[CutResult](c, ctx, "goposix.cut", map[string]interface{}{"flags": flags})
}

// UniqItem is a single item from uniq --json output.
type UniqItem struct {
	Line  string `json:"line"`
	Count int    `json:"count"`
}

// Uniq runs goposix.uniq. Returns items directly (uniq output is an array).
func (c *Client) Uniq(ctx context.Context, flags []string) ([]UniqItem, error) {
	var raw daemonResult
	if err := c.Call(ctx, "goposix.uniq", map[string]interface{}{"flags": flags}, &raw); err != nil {
		return nil, err
	}
	var items []UniqItem
	if raw.Data != nil {
		if err := json.Unmarshal(raw.Data, &items); err != nil {
			return nil, err
		}
	}
	return items, nil
}

// GrepMatch is a single match from grep --json output.
type GrepMatch struct {
	File    string   `json:"file,omitempty"`
	Line    int      `json:"line"`
	Text    string   `json:"text"`
	Matches []string `json:"matches"`
}

// Grep runs goposix.grep. pattern is always the first positional argument.
func (c *Client) Grep(ctx context.Context, pattern string, flags []string) ([]GrepMatch, error) {
	allFlags := append([]string{pattern}, flags...)
	var raw daemonResult
	if err := c.Call(ctx, "goposix.grep", map[string]interface{}{"flags": allFlags}, &raw); err != nil {
		return nil, err
	}
	var matches []GrepMatch
	if raw.Data != nil {
		if err := json.Unmarshal(raw.Data, &matches); err != nil {
			return nil, err
		}
	}
	return matches, nil
}

// WcResult is the output of wc --json (single file).
type WcResult struct {
	Lines int `json:"lines"`
	Words int `json:"words"`
	Bytes int `json:"bytes"`
	Chars int `json:"chars"`
}

// Wc runs goposix.wc.
func (c *Client) Wc(ctx context.Context, path string) (*WcResult, error) {
	var raw daemonResult
	if err := c.Call(ctx, "goposix.wc", map[string]string{"path": path}, &raw); err != nil {
		return nil, err
	}
	var data WcResult
	if raw.Data != nil {
		if err := json.Unmarshal(raw.Data, &data); err != nil {
			return nil, err
		}
	}
	return &data, nil
}

// --- File stat ---

// StatResult is the output of stat --json.
type StatResult struct {
	Path   string `json:"path"`
	Size   int64  `json:"size"`
	Mode   string `json:"mode"`
	UID    uint32 `json:"uid"`
	GID    uint32 `json:"gid"`
	Mtime  string `json:"mtime"`
	Inode  uint64 `json:"inode"`
	Links  uint64 `json:"links"`
	Blocks int64  `json:"blocks"`
	IsDir  bool   `json:"isDir"`
	IsLink bool   `json:"isLink"`
}

// Stat runs goposix.stat.
func (c *Client) Stat(ctx context.Context, path string) (*StatResult, error) {
	return callUtility[StatResult](c, ctx, "goposix.stat", map[string]string{"path": path})
}

// --- File listing ---

// FileInfo is a single file entry from ls --json.
type FileInfo struct {
	Name    string `json:"name"`
	Path    string `json:"path"`
	Size    int64  `json:"size"`
	Mode    string `json:"mode"`
	ModTime string `json:"modTime"`
	IsDir   bool   `json:"isDir"`
	Owner   string `json:"owner"`
	Group   string `json:"group"`
	Inode   uint64 `json:"inode"`
	Links   uint64 `json:"links"`
	Target  string `json:"target,omitempty"`
	Blocks  int64  `json:"blocks"`
}

// LsResult is the output of ls --json.
type LsResult struct {
	Path  string     `json:"path"`
	Files []FileInfo `json:"files"`
	Total int        `json:"total"`
}

// Ls runs goposix.ls.
func (c *Client) Ls(ctx context.Context, path string, flags []string) (*LsResult, error) {
	return callUtility[LsResult](c, ctx, "goposix.ls", map[string]interface{}{"path": path, "flags": flags})
}

// FindEntry is a single entry from find --json output.
type FindEntry struct {
	Path  string `json:"path"`
	Type  string `json:"type"`
	Size  int64  `json:"size"`
	Mtime string `json:"mtime"`
}

// Find runs goposix.find. Returns entries directly (find output is an array).
func (c *Client) Find(ctx context.Context, basePath string, flags []string) ([]FindEntry, error) {
	var raw daemonResult
	params := map[string]interface{}{"flags": flags}
	if basePath != "" {
		params["path"] = basePath
	}
	if err := c.Call(ctx, "goposix.find", params, &raw); err != nil {
		return nil, err
	}
	var entries []FindEntry
	if raw.Data != nil {
		if err := json.Unmarshal(raw.Data, &entries); err != nil {
			return nil, err
		}
	}
	return entries, nil
}

// --- File operations ---

// MoveRecord is a single entry in mv --json output.
type MoveRecord struct {
	From string `json:"from"`
	To   string `json:"to"`
}

// MvResult is the output of mv --json.
type MvResult struct {
	Moved []MoveRecord `json:"moved"`
}

// Mv runs goposix.mv.
func (c *Client) Mv(ctx context.Context, from, to string) (*MvResult, error) {
	return callUtility[MvResult](c, ctx, "goposix.mv", map[string]interface{}{"flags": []string{from, to}})
}

// CopyRecord is a single entry in cp --json output.
type CopyRecord struct {
	From string `json:"from"`
	To   string `json:"to"`
}

// CpResult is the output of cp --json.
type CpResult struct {
	Copied []CopyRecord `json:"copied"`
}

// Cp runs goposix.cp.
func (c *Client) Cp(ctx context.Context, from, to string) (*CpResult, error) {
	return callUtility[CpResult](c, ctx, "goposix.cp", map[string]interface{}{"flags": []string{from, to}})
}

// LnResult is the output of ln --json.
type LnResult struct {
	Links []LinkEntry `json:"links"`
}

// LinkEntry is a single entry in ln --json output.
type LinkEntry struct {
	Target string `json:"target"`
	Link   string `json:"link"`
}

// Ln runs goposix.ln.
func (c *Client) Ln(ctx context.Context, target, link string, symbolic bool) (*LnResult, error) {
	flags := []string{target, link}
	if symbolic {
		flags = append([]string{"-s"}, flags...)
	}
	return callUtility[LnResult](c, ctx, "goposix.ln", map[string]interface{}{"flags": flags})
}

// RmResult is the output of rm --json.
type RmResult struct {
	Removed []string `json:"removed"`
	Errors  []string `json:"errors,omitempty"`
}

// Rm runs goposix.rm.
func (c *Client) Rm(ctx context.Context, paths []string, recursive, force bool) (*RmResult, error) {
	flags := paths
	if recursive {
		flags = append([]string{"-r"}, flags...)
	}
	if force {
		flags = append([]string{"-f"}, flags...)
	}
	return callUtility[RmResult](c, ctx, "goposix.rm", map[string]interface{}{"flags": flags})
}

// RmdirResult is the output of rmdir --json.
type RmdirResult struct {
	Removed []string `json:"removed"`
}

// Rmdir runs goposix.rmdir.
func (c *Client) Rmdir(ctx context.Context, path string) (*RmdirResult, error) {
	return callUtility[RmdirResult](c, ctx, "goposix.rmdir", map[string]string{"path": path})
}

// MkdirResult is the output of mkdir --json.
type MkdirResult struct {
	Created []string `json:"created"`
}

// Mkdir runs goposix.mkdir. If parents is true, creates parent directories.
func (c *Client) Mkdir(ctx context.Context, path string, parents bool) (*MkdirResult, error) {
	flags := []string{}
	if parents {
		flags = append(flags, "-p")
	}
	flags = append(flags, path)
	return callUtility[MkdirResult](c, ctx, "goposix.mkdir", map[string]interface{}{"flags": flags})
}

// TouchResult is the output of touch --json.
type TouchResult struct {
	Touched []string `json:"touched"`
}

// Touch runs goposix.touch.
func (c *Client) Touch(ctx context.Context, paths []string) (*TouchResult, error) {
	return callUtility[TouchResult](c, ctx, "goposix.touch", map[string]interface{}{"flags": paths})
}

// ChmodRecord is a single entry in chmod --json output.
type ChmodRecord struct {
	Path string `json:"path"`
	Mode string `json:"mode"`
}

// ChmodResult is the output of chmod --json.
type ChmodResult struct {
	Changed []ChmodRecord `json:"changed"`
}

// Chmod runs goposix.chmod.
func (c *Client) Chmod(ctx context.Context, mode string, paths []string) (*ChmodResult, error) {
	flags := append([]string{mode}, paths...)
	return callUtility[ChmodResult](c, ctx, "goposix.chmod", map[string]interface{}{"flags": flags})
}

// ChownRecord is a single entry in chown/chgrp --json output.
type ChownRecord struct {
	Path string `json:"path"`
}

// ChownResult is the output of chown/chgrp --json.
type ChownResult struct {
	Changed []ChownRecord `json:"changed"`
}

// Chown runs goposix.chown.
func (c *Client) Chown(ctx context.Context, owner string, paths []string) (*ChownResult, error) {
	flags := append([]string{owner}, paths...)
	return callUtility[ChownResult](c, ctx, "goposix.chown", map[string]interface{}{"flags": flags})
}

// Chgrp runs goposix.chgrp.
func (c *Client) Chgrp(ctx context.Context, group string, paths []string) (*ChownResult, error) {
	flags := append([]string{group}, paths...)
	return callUtility[ChownResult](c, ctx, "goposix.chgrp", map[string]interface{}{"flags": flags})
}

// --- Hash utilities ---

// HashEntry is a single entry from md5sum/sha256sum --json (hash mode).
type HashEntry struct {
	File      string `json:"file"`
	Hash      string `json:"hash"`
	Algorithm string `json:"algorithm"`
}

// CheckEntry is a single entry from md5sum/sha256sum --json (check mode).
type CheckEntry struct {
	File   string `json:"file"`
	Status string `json:"status"`
}

// Md5sum runs goposix.md5sum.
func (c *Client) Md5sum(ctx context.Context, paths []string, check bool) (json.RawMessage, error) {
	flags := paths
	if check {
		flags = append([]string{"-c"}, flags...)
	}
	var raw daemonResult
	if err := c.Call(ctx, "goposix.md5sum", map[string]interface{}{"flags": flags}, &raw); err != nil {
		return nil, err
	}
	return raw.Data, nil
}

// Sha256sum runs goposix.sha256sum.
func (c *Client) Sha256sum(ctx context.Context, paths []string, check bool) (json.RawMessage, error) {
	flags := paths
	if check {
		flags = append([]string{"-c"}, flags...)
	}
	var raw daemonResult
	if err := c.Call(ctx, "goposix.sha256sum", map[string]interface{}{"flags": flags}, &raw); err != nil {
		return nil, err
	}
	return raw.Data, nil
}

// --- Archive ---

// GzipStat is a single entry from gzip --json output.
type GzipStat struct {
	File         string  `json:"file"`
	OriginalSize int64   `json:"originalSize"`
	NewSize      int64   `json:"newSize"`
	Ratio        float64 `json:"ratio"`
}

// Gzip runs goposix.gzip.
func (c *Client) Gzip(ctx context.Context, flags []string) ([]GzipStat, error) {
	var raw daemonResult
	if err := c.Call(ctx, "goposix.gzip", map[string]interface{}{"flags": flags}, &raw); err != nil {
		return nil, err
	}
	var stats []GzipStat
	if raw.Data != nil {
		if err := json.Unmarshal(raw.Data, &stats); err != nil {
			return nil, err
		}
	}
	return stats, nil
}

// TarFileStat is a single entry from tar --json output.
type TarFileStat struct {
	Name string `json:"name"`
	Size int64  `json:"size"`
	Mode string `json:"mode"`
}

// Tar runs goposix.tar.
func (c *Client) Tar(ctx context.Context, flags []string) ([]TarFileStat, error) {
	var raw daemonResult
	if err := c.Call(ctx, "goposix.tar", map[string]interface{}{"flags": flags}, &raw); err != nil {
		return nil, err
	}
	var stats []TarFileStat
	if raw.Data != nil {
		if err := json.Unmarshal(raw.Data, &stats); err != nil {
			return nil, err
		}
	}
	return stats, nil
}

// --- Filesystem ---

// FSInfo is a single entry from df --json output.
type FSInfo struct {
	Filesystem string `json:"filesystem"`
	Size       uint64 `json:"size"`
	Used       uint64 `json:"used"`
	Avail      uint64 `json:"avail"`
	Mountpoint string `json:"mountpoint"`
}

// Df runs goposix.df.
func (c *Client) Df(ctx context.Context, path string) ([]FSInfo, error) {
	var raw daemonResult
	params := map[string]interface{}{}
	if path != "" {
		params["path"] = path
	}
	if err := c.Call(ctx, "goposix.df", params, &raw); err != nil {
		return nil, err
	}
	var infos []FSInfo
	if raw.Data != nil {
		if err := json.Unmarshal(raw.Data, &infos); err != nil {
			return nil, err
		}
	}
	return infos, nil
}

// DirInfo is a single entry from du --json output.
type DirInfo struct {
	Path  string `json:"path"`
	Size  int64  `json:"size"`
	Files int    `json:"files"`
}

// Du runs goposix.du.
func (c *Client) Du(ctx context.Context, path string) ([]DirInfo, error) {
	var raw daemonResult
	if err := c.Call(ctx, "goposix.du", map[string]string{"path": path}, &raw); err != nil {
		return nil, err
	}
	var infos []DirInfo
	if raw.Data != nil {
		if err := json.Unmarshal(raw.Data, &infos); err != nil {
			return nil, err
		}
	}
	return infos, nil
}

// --- Process ---

// ProcessInfo is a single entry from ps --json output.
type ProcessInfo struct {
	PID  int    `json:"pid"`
	PPID int    `json:"ppid"`
	User string `json:"user"`
	Cmd  string `json:"cmd"`
	CPU  string `json:"cpu"`
	Mem  string `json:"mem"`
}

// Ps runs goposix.ps.
func (c *Client) Ps(ctx context.Context) ([]ProcessInfo, error) {
	var raw daemonResult
	if err := c.Call(ctx, "goposix.ps", nil, &raw); err != nil {
		return nil, err
	}
	var procs []ProcessInfo
	if raw.Data != nil {
		if err := json.Unmarshal(raw.Data, &procs); err != nil {
			return nil, err
		}
	}
	return procs, nil
}

// KillRecord is a single entry in kill --json output.
type KillRecord struct {
	PID     int    `json:"pid"`
	Signal  string `json:"signal"`
	Success bool   `json:"success"`
}

// KillResult is the output of kill --json.
type KillResult struct {
	Signaled []KillRecord `json:"signaled"`
}

// Kill runs goposix.kill.
func (c *Client) Kill(ctx context.Context, signal string, pids []int) (*KillResult, error) {
	pidStrs := make([]string, len(pids))
	for i, p := range pids {
		pidStrs[i] = itoa(p)
	}
	flags := pidStrs
	if signal != "" {
		flags = append([]string{"-s", signal}, flags...)
	}
	return callUtility[KillResult](c, ctx, "goposix.kill", map[string]interface{}{"flags": flags})
}

// --- Diff ---

// Hunk is a single hunk in diff --json output.
type Hunk struct {
	OldStart int      `json:"oldStart"`
	OldLines int      `json:"oldLines"`
	NewStart int      `json:"newStart"`
	NewLines int      `json:"newLines"`
	Lines    []string `json:"lines"`
}

// DiffResult is the output of diff --json.
type DiffResult struct {
	Files  []string `json:"files"`
	Differ bool     `json:"differ"`
	Hunks  []Hunk   `json:"hunks"`
}

// Diff runs goposix.diff.
func (c *Client) Diff(ctx context.Context, file1, file2 string) (*DiffResult, error) {
	return callUtility[DiffResult](c, ctx, "goposix.diff", map[string]interface{}{"flags": []string{file1, file2}})
}

// --- Command execution ---

// ExecEntry is a single entry from xargs --json output.
type ExecEntry struct {
	Command  string `json:"command"`
	ExitCode int    `json:"exitCode"`
}

// Xargs runs goposix.xargs.
func (c *Client) Xargs(ctx context.Context, command string, flags []string) ([]ExecEntry, error) {
	var raw daemonResult
	params := map[string]interface{}{
		"flags": append(flags, command),
	}
	if err := c.Call(ctx, "goposix.xargs", params, &raw); err != nil {
		return nil, err
	}
	var entries []ExecEntry
	if raw.Data != nil {
		if err := json.Unmarshal(raw.Data, &entries); err != nil {
			return nil, err
		}
	}
	return entries, nil
}

// ExprResult is the output of expr --json.
type ExprResult struct {
	Result   string `json:"result"`
	ExitCode int    `json:"exitCode"`
}

// Expr runs goposix.expr.
func (c *Client) Expr(ctx context.Context, expression []string) (*ExprResult, error) {
	return callUtility[ExprResult](c, ctx, "goposix.expr", map[string]interface{}{"flags": expression})
}

// --- Session management ---

// SessionInfo represents a goposix session.
type SessionInfo struct {
	SessionID  string            `json:"sessionId"`
	CWD        string            `json:"cwd"`
	Env        map[string]string `json:"env"`
	LastActive string            `json:"lastActive"`
}

// SessionCreate creates a new session.
func (c *Client) SessionCreate(ctx context.Context) (*SessionInfo, error) {
	var s SessionInfo
	if err := c.Call(ctx, "goposix.session.create", nil, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

// SessionSetCwd sets the working directory for a session.
func (c *Client) SessionSetCwd(ctx context.Context, sessionID, path string) error {
	var ok bool
	if err := c.Call(ctx, "goposix.session.setCwd", map[string]string{
		"sessionId": sessionID,
		"path":      path,
	}, &ok); err != nil {
		return err
	}
	return nil
}

// SessionList returns all active sessions.
func (c *Client) SessionList(ctx context.Context) ([]SessionInfo, error) {
	var sessions []SessionInfo
	if err := c.Call(ctx, "goposix.session.list", nil, &sessions); err != nil {
		return nil, err
	}
	return sessions, nil
}

// SessionDestroy destroys a session.
func (c *Client) SessionDestroy(ctx context.Context, sessionID string) error {
	var ok bool
	if err := c.Call(ctx, "goposix.session.destroy", map[string]string{
		"sessionId": sessionID,
	}, &ok); err != nil {
		return err
	}
	return nil
}

// --- Shell execution ---

// ExecResult is the output of goposix.shell.exec.
type ExecResult struct {
	Stdout   string `json:"stdout"`
	Stderr   string `json:"stderr"`
	ExitCode uint8  `json:"exitCode"`
}

// ShellExec runs a shell script in the given session.
func (c *Client) ShellExec(ctx context.Context, sessionID, script string) (*ExecResult, error) {
	var res ExecResult
	if err := c.Call(ctx, "goposix.shell.exec", map[string]string{
		"sessionId": sessionID,
		"script":    script,
	}, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

// --- Ping ---

// PingResult is the output of goposix.ping.
type PingResult struct {
	Pong    bool   `json:"pong"`
	Uptime  string `json:"uptime"`
	Version string `json:"version"`
}

// Ping checks daemon health.
func (c *Client) Ping(ctx context.Context) (*PingResult, error) {
	var p PingResult
	if err := c.Call(ctx, "goposix.ping", nil, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := false
	if n < 0 {
		neg = true
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
