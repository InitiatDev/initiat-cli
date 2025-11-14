package setup

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

type ValidationErrors []ValidationError

func (e ValidationErrors) Error() string {
	if len(e) == 0 {
		return ""
	}

	var messages []string
	for _, err := range e {
		messages = append(messages, err.Error())
	}
	return strings.Join(messages, "; ")
}

func Validate(config *SetupConfig) error {
	var errors ValidationErrors

	errors = append(errors, validateVersion(config)...)
	errors = append(errors, validateSteps(config)...)
	errors = append(errors, validateSecrets(config)...)
	errors = append(errors, validateTimeouts(config)...)
	errors = append(errors, validatePaths(config)...)

	if len(errors) > 0 {
		return errors
	}

	return nil
}

func validateVersion(config *SetupConfig) ValidationErrors {
	var errors ValidationErrors

	if config.Version != 1 {
		errors = append(errors, ValidationError{
			Field:   "version",
			Message: fmt.Sprintf("must be 1, got %d", config.Version),
		})
	}

	return errors
}

func validateSteps(config *SetupConfig) ValidationErrors {
	var errors ValidationErrors

	for _, phase := range GetAllPhases(config) {
		for i, step := range phase.Steps {
			stepErrors := validateStep(step, fmt.Sprintf("%s[%d]", phase.Name, i))
			errors = append(errors, stepErrors...)
		}
	}

	return errors
}

func validateStep(step Step, path string) ValidationErrors {
	var errors ValidationErrors

	errors = append(errors, validateStepActions(step, path)...)
	errors = append(errors, validateStepTimeout(step, path)...)
	errors = append(errors, validateStepRetries(step, path)...)

	return errors
}

func validateStepActions(step Step, path string) ValidationErrors {
	var errors ValidationErrors

	actionCount := 0
	actionFields := []string{}

	if step.Run != "" {
		actionCount++
		actionFields = append(actionFields, "run")
	}
	if step.Print != "" {
		actionCount++
		actionFields = append(actionFields, "print")
	}

	if actionCount == 0 {
		errors = append(errors, ValidationError{
			Field:   path,
			Message: "step must have exactly one action (run or print)",
		})
	} else if actionCount > 1 {
		errors = append(errors, ValidationError{
			Field:   path,
			Message: fmt.Sprintf("step has multiple actions: %s", strings.Join(actionFields, ", ")),
		})
	}

	return errors
}

func validateStepTimeout(step Step, path string) ValidationErrors {
	var errors ValidationErrors

	if step.Timeout != "" {
		if err := validateDuration(step.Timeout); err != nil {
			errors = append(errors, ValidationError{
				Field:   fmt.Sprintf("%s.timeout", path),
				Message: err.Error(),
			})
		}
	}

	return errors
}

func validateStepRetries(step Step, path string) ValidationErrors {
	var errors ValidationErrors

	if step.Retries != nil {
		if step.Retries.Attempts < 1 {
			errors = append(errors, ValidationError{
				Field:   fmt.Sprintf("%s.retries.attempts", path),
				Message: "must be at least 1",
			})
		}
		if err := validateDuration(step.Retries.Backoff); err != nil {
			errors = append(errors, ValidationError{
				Field:   fmt.Sprintf("%s.retries.backoff", path),
				Message: err.Error(),
			})
		}
	}

	return errors
}

func validateSecrets(config *SetupConfig) ValidationErrors {
	var errors ValidationErrors

	if config.Env == nil {
		return errors
	}

	declaredSecrets := make(map[string]bool)
	for _, secret := range config.Env.Secrets {
		declaredSecrets[secret] = true
	}

	for _, phase := range GetAllPhases(config) {
		for i, step := range phase.Steps {
			for j, secret := range step.EnvFromSecrets {
				if !declaredSecrets[secret] {
					errors = append(errors, ValidationError{
						Field:   fmt.Sprintf("%s[%d].env_from_secrets[%d]", phase.Name, i, j),
						Message: fmt.Sprintf("secret '%s' not declared in env.secrets", secret),
					})
				}
			}
		}
	}

	return errors
}

func validateTimeouts(config *SetupConfig) ValidationErrors {
	var errors ValidationErrors

	if config.Defaults != nil && config.Defaults.Timeout != "" {
		if err := validateDuration(config.Defaults.Timeout); err != nil {
			errors = append(errors, ValidationError{
				Field:   "defaults.timeout",
				Message: err.Error(),
			})
		}
	}

	return errors
}

func validatePaths(config *SetupConfig) ValidationErrors {
	var errors ValidationErrors

	for _, phase := range GetAllPhases(config) {
		for i, step := range phase.Steps {
			if step.CWD != "" {
				if filepath.IsAbs(step.CWD) {
					errors = append(errors, ValidationError{
						Field:   fmt.Sprintf("%s[%d].cwd", phase.Name, i),
						Message: "must be relative path (no absolute paths allowed)",
					})
				}
				if strings.Contains(step.CWD, "..") {
					errors = append(errors, ValidationError{
						Field:   fmt.Sprintf("%s[%d].cwd", phase.Name, i),
						Message: "must not contain '..' (no parent directory access)",
					})
				}
			}
		}
	}

	return errors
}

func validateDuration(duration string) error {
	matched, _ := regexp.MatchString(`^\d+[smhd]$`, duration)
	if !matched {
		return fmt.Errorf("must be in format '15m', '2h', etc")
	}

	_, err := time.ParseDuration(duration)
	if err != nil {
		return fmt.Errorf("invalid duration: %w", err)
	}

	return nil
}
