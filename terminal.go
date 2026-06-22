package kvm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/user"
	"strconv"
	"syscall"

	"github.com/creack/pty"
	"github.com/pion/webrtc/v4"
)

type TerminalSize struct {
	Rows int `json:"rows"`
	Cols int `json:"cols"`
}

// resolveShell returns the path to the preferred available shell, trying
// zsh, then bash, then sh.
func resolveShell() (string, error) {
	for _, name := range []string{"zsh", "bash", "sh"} {
		if path, err := exec.LookPath(name); err == nil {
			return path, nil
		}
	}
	return "", fmt.Errorf("no shell found (tried zsh, bash, sh)")
}

// unprivilegedCredential resolves the "nobody" account and returns a
// syscall.Credential that drops the spawned shell to its uid/gid and clears
// root's supplementary groups. The daemon itself stays root for hardware
// access; only the interactive shell is de-privileged.
func unprivilegedCredential() (*syscall.Credential, error) {
	u, err := user.Lookup("nobody")
	if err != nil {
		return nil, fmt.Errorf("lookup nobody: %w", err)
	}
	uid, err := strconv.ParseUint(u.Uid, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("parse nobody uid %q: %w", u.Uid, err)
	}
	gid, err := strconv.ParseUint(u.Gid, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("parse nobody gid %q: %w", u.Gid, err)
	}
	return &syscall.Credential{
		Uid:    uint32(uid),
		Gid:    uint32(gid),
		Groups: []uint32{}, // drop root's supplementary groups
	}, nil
}

func handleTerminalChannel(d *webrtc.DataChannel) {
	scopedLogger := terminalLogger.With().
		Uint16("data_channel_id", *d.ID()).Logger()

	var ptmx *os.File
	var cmd *exec.Cmd
	d.OnOpen(func() {
		// Prefer zsh, then bash, then sh.
		shellPath, err := resolveShell()
		if err != nil {
			scopedLogger.Error().Err(err).Msg("Refusing to start terminal: no shell available")
			d.Close()
			return
		}
		cmd = exec.Command(shellPath)

		// The KVM daemon runs as root for hardware access, but the interactive
		// shell does not need privileges. Drop to the unprivileged "nobody"
		// user before exec so a compromised terminal can't own the device. Fail
		// closed rather than hand out a root shell if "nobody" can't be resolved.
		cred, err := unprivilegedCredential()
		if err != nil {
			scopedLogger.Error().Err(err).Msg("Refusing to start terminal: cannot drop privileges")
			d.Close()
			return
		}
		cmd.SysProcAttr = &syscall.SysProcAttr{Credential: cred}

		// Don't leak the daemon's environment into the shell. nobody has no home,
		// so point HOME at world-writable /tmp. xterm.js identifies as an xterm
		// terminal; without TERM, shells that rely on terminfo for key bindings
		// (e.g. zsh's ZLE) mis-handle keys like Backspace.
		cmd.Env = []string{
			"TERM=xterm-256color",
			"HOME=/tmp",
			"PATH=/usr/bin:/bin",
			"USER=nobody",
			"SHELL=" + shellPath,
		}
		cmd.Dir = "/tmp"

		ptmx, err = pty.Start(cmd)
		if err != nil {
			scopedLogger.Warn().Err(err).Msg("Failed to start pty")
			d.Close()
			return
		}

		go func() {
			buf := make([]byte, 1024)
			for {
				n, err := ptmx.Read(buf)
				if err != nil {
					if err != io.EOF {
						scopedLogger.Warn().Err(err).Msg("Failed to read from pty")
					}
					break
				}
				err = d.Send(buf[:n])
				if err != nil {
					scopedLogger.Warn().Err(err).Msg("Failed to send pty output")
					break
				}
			}
		}()
	})

	d.OnMessage(func(msg webrtc.DataChannelMessage) {
		if ptmx == nil {
			return
		}
		if msg.IsString {
			maybeJson := bytes.TrimSpace(msg.Data)
			// Cheap check to see if this resembles JSON
			if len(maybeJson) > 1 && maybeJson[0] == '{' && maybeJson[len(maybeJson)-1] == '}' {
				var size TerminalSize
				err := json.Unmarshal(maybeJson, &size)
				if err == nil {
					err = pty.Setsize(ptmx, &pty.Winsize{
						Rows: uint16(size.Rows),
						Cols: uint16(size.Cols),
					})
					if err == nil {
						scopedLogger.Info().Int("rows", size.Rows).Int("cols", size.Cols).Msg("Set terminal size")
						return
					}
				}
				scopedLogger.Warn().Err(err).Msg("Failed to parse terminal size")
			}
		}
		_, err := ptmx.Write(msg.Data)
		if err != nil {
			scopedLogger.Warn().Err(err).Msg("Failed to write to pty")
		}
	})

	d.OnClose(func() {
		if ptmx != nil {
			ptmx.Close()
		}
		if cmd != nil && cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
		scopedLogger.Info().Msg("Terminal channel closed")
	})

	d.OnError(func(err error) {
		scopedLogger.Warn().Err(err).Msg("Terminal channel error")
	})
}
