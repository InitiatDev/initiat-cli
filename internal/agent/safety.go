package agent

import (
	"path/filepath"
	"regexp"
	"strings"
)

type SafetyAssessment struct {
	EffectiveDanger DangerLevel
	Signals         []string
}

func AssessSafety(action ProposedAction) SafetyAssessment {
	if action.Type == ActionRunCommand {
		return assessCommandSafety(action.Command)
	}
	if action.Type == ActionEditFiles {
		return assessEditSafety(action.Edits)
	}
	return SafetyAssessment{EffectiveDanger: DangerSafe}
}

func MaxDanger(a, b DangerLevel) DangerLevel {
	if dangerRank(a) >= dangerRank(b) {
		return a
	}
	return b
}

func dangerRank(d DangerLevel) int {
	const (
		rankSafe      = 0
		rankCaution   = 1
		rankDangerous = 2
	)

	switch d {
	case DangerSafe:
		return rankSafe
	case DangerCaution:
		return rankCaution
	case DangerDangerous:
		return rankDangerous
	default:
		return rankDangerous
	}
}

var (
	reWordSudo  = regexp.MustCompile(`(?i)(^|[^a-z0-9_])sudo([^a-z0-9_]|$)`)
	reWordRm    = regexp.MustCompile(`(?i)(^|[^a-z0-9_])rm([^a-z0-9_]|$)`)
	reWordChmod = regexp.MustCompile(`(?i)(^|[^a-z0-9_])chmod([^a-z0-9_]|$)`)
	reWordChown = regexp.MustCompile(`(?i)(^|[^a-z0-9_])chown([^a-z0-9_]|$)`)
	reWordChgrp = regexp.MustCompile(`(?i)(^|[^a-z0-9_])chgrp([^a-z0-9_]|$)`)
	rePipeToSh  = regexp.MustCompile(`(?is)\b(curl|wget)\b[\s\S]*\|\s*(sh|bash|zsh)\b`)

	reLikelyMutating = regexp.MustCompile(`(?i)\b(` +
		`bundle\s+install|` +
		`bundle\s+exec\s+rails\s+db:(migrate|schema:load|setup|prepare)|` +
		`rails\s+db:(migrate|schema:load|setup|prepare)|` +
		`rake\s+db:(migrate|schema:load|setup|prepare)|` +
		`npm\s+(ci|install)|` +
		`pnpm\s+(i|install)|` +
		`yarn(\s+install)?|` +
		`bun\s+install|` +
		`go\s+mod\s+(tidy|download)|` +
		`pip\s+install|` +
		`poetry\s+install|` +
		`mix\s+deps\.get|` +
		`gem\s+install|` +
		`brew\s+install|` +
		`asdf\s+install|` +
		`mise\s+install` +
		`)\b`)
)

func assessCommandSafety(command string) SafetyAssessment {
	c := strings.TrimSpace(command)
	if c == "" {
		return SafetyAssessment{EffectiveDanger: DangerSafe}
	}

	d := DangerSafe
	var signals []string

	if reWordSudo.MatchString(c) {
		d = MaxDanger(d, DangerDangerous)
		signals = append(signals, "uses sudo (privilege escalation)")
	}
	if rePipeToSh.MatchString(c) {
		d = MaxDanger(d, DangerDangerous)
		signals = append(signals, "pipes network content to a shell")
	}
	if reWordRm.MatchString(c) {
		d = MaxDanger(d, DangerDangerous)
		signals = append(signals, "deletes files (rm)")
	}
	if reWordChmod.MatchString(c) || reWordChown.MatchString(c) || reWordChgrp.MatchString(c) {
		d = MaxDanger(d, DangerDangerous)
		signals = append(signals, "changes permissions/ownership (chmod/chown/chgrp)")
	}
	if strings.Contains(c, ">") {
		d = MaxDanger(d, DangerCaution)
		signals = append(signals, "uses shell redirection (may overwrite files)")
	}
	if strings.Contains(strings.ToLower(c), "git clean") {
		d = MaxDanger(d, DangerDangerous)
		signals = append(signals, "may delete untracked files (git clean)")
	}

	return SafetyAssessment{EffectiveDanger: d, Signals: signals}
}

func CommandLikelyMutatesProject(command string) bool {
	c := strings.TrimSpace(command)
	if c == "" {
		return false
	}
	return reLikelyMutating.MatchString(c)
}

func assessEditSafety(edits []FileEdit) SafetyAssessment {
	d := DangerSafe
	var signals []string

	for _, e := range edits {
		p := strings.TrimSpace(e.Path)
		if p == "" {
			continue
		}
		clean := filepath.Clean(p)
		lower := strings.ToLower(clean)

		if strings.HasPrefix(clean, string(filepath.Separator)) {
			d = MaxDanger(d, DangerDangerous)
			signals = append(signals, "edits an absolute path (possible system config)")
		}

		parts := strings.Split(lower, string(filepath.Separator))
		if containsPart(parts, ".ssh") {
			d = MaxDanger(d, DangerDangerous)
			signals = append(signals, "edits SSH-related files")
		}

		base := strings.ToLower(filepath.Base(clean))
		switch {
		case strings.HasPrefix(base, ".env"):
			d = MaxDanger(d, DangerDangerous)
			signals = append(signals, "edits .env-style secret file")
		case base == ".envrc":
			d = MaxDanger(d, DangerDangerous)
			signals = append(signals, "edits direnv config (.envrc)")
		case base == ".gitconfig":
			d = MaxDanger(d, DangerDangerous)
			signals = append(signals, "edits git config")
		case base == ".zshrc" || base == ".bashrc" || base == ".bash_profile" || base == ".profile":
			d = MaxDanger(d, DangerDangerous)
			signals = append(signals, "edits shell startup config")
		}
	}

	return SafetyAssessment{EffectiveDanger: d, Signals: uniqueStrings(signals)}
}

func containsPart(parts []string, want string) bool {
	for _, p := range parts {
		if p == want {
			return true
		}
	}
	return false
}

func uniqueStrings(in []string) []string {
	if len(in) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(in))
	out := make([]string, 0, len(in))
	for _, s := range in {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	return out
}
