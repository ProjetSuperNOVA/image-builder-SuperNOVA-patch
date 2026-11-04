package osbuild

import (
	"fmt"
	"strings"
)

// Grub2DStageOptions represents options for the
// org.osbuild.grub2.d stage.
//
// This stage writes a GRUB2 drop-in configuration file at a
// configurable path relative to the filesystem root.
type Grub2DStageOptions struct {
	// Path relative to the filesystem root
	Path string `json:"path"`

	// GRUB2 configuration to write
	Config *Grub2DConfig `json:"config"`
}

// Grub2DConfig contains GRUB2 settings for a drop-in config file.
type Grub2DConfig struct {
	TerminalInput  []string `json:"terminal_input,omitempty"`
	TerminalOutput []string `json:"terminal_output,omitempty"`
	Serial         string   `json:"serial,omitempty"`
}

func (Grub2DStageOptions) isStageOptions() {}

func NewGrub2DStage(options *Grub2DStageOptions) *Stage {
	return &Stage{
		Type:    "org.osbuild.grub2.d",
		Options: options,
	}
}

// NewGrub2DConfigFromGrub2Config creates a Grub2DConfig from a
// GRUB2Config, extracting only the console-related fields.
// Returns nil if no console settings are present.
func NewGrub2DConfigFromGrub2Config(cfg *GRUB2Config) *Grub2DConfig {
	if cfg == nil {
		return nil
	}
	c := &Grub2DConfig{
		TerminalInput:  cfg.TerminalInput,
		TerminalOutput: cfg.TerminalOutput,
		Serial:         cfg.Serial,
	}
	if len(c.TerminalInput) == 0 && len(c.TerminalOutput) == 0 && c.Serial == "" {
		return nil
	}
	return c
}

func ValidateGRUB2TimeoutStyle(style GRUB2ConfigTimeoutStyle) error {
	switch style {
	case "", GRUB2ConfigTimeoutStyleCountdown, GRUB2ConfigTimeoutStyleHidden, GRUB2ConfigTimeoutStyleMenu:
		return nil
	default:
		return fmt.Errorf("invalid GRUB2 timeout style %q", style)
	}
}

// RenderBootcGRUB2Config renders the drop-in consumed by Fedora's bootc GRUB
// configuration. The terminal removals are intentionally separate from menu
// suppression: timeout settings do not suppress GRUB output.
func RenderBootcGRUB2Config(cfg *GRUB2Config) (string, error) {
	if cfg == nil {
		cfg = &GRUB2Config{}
	}
	if err := ValidateGRUB2TimeoutStyle(cfg.TimeoutStyle); err != nil {
		return "", err
	}

	lines := []string{
		"set pager=0",
		"set timeout=0",
		"set timeout_style=hidden",
		"terminal_input --remove console",
		"terminal_output --remove console gfxterm vga_text",
	}
	if cfg.Serial != "" {
		lines = append(lines, cfg.Serial)
	}
	if terminals := nonConsoleTerminals(cfg.TerminalInput); len(terminals) > 0 {
		lines = append(lines, "terminal_input "+strings.Join(terminals, " "))
	}
	if terminals := nonConsoleTerminals(cfg.TerminalOutput); len(terminals) > 0 {
		lines = append(lines, "terminal_output "+strings.Join(terminals, " "))
	}
	return strings.Join(lines, "\n") + "\n", nil
}

func nonConsoleTerminals(terminals []string) []string {
	result := make([]string, 0, len(terminals))
	for _, terminal := range terminals {
		switch terminal {
		case "--remove", "console", "gfxterm", "vga_text", "mda_text":
			continue
		default:
			result = append(result, terminal)
		}
	}
	return result
}
