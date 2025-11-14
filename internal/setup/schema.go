package setup

import (
	"encoding/json"
	"fmt"
	"os"
)

const (
	httpStatusMin     = 100
	httpStatusMax     = 599
	fileModeReadWrite = 0600
)

func GenerateJSONSchema() ([]byte, error) {
	schema := map[string]interface{}{
		"$schema":     "https://json-schema.org/draft/2020-12/schema",
		"$id":         "https://initiat.dev/schemas/setup-v1.json",
		"title":       "Initiat Setup Configuration",
		"description": "Configuration schema for .initiat/setup.yml files",
		"type":        "object",
		"required":    []string{"version"},
		"properties": map[string]interface{}{
			"version": map[string]interface{}{
				"type":        "integer",
				"const":       1,
				"description": "Schema version (must be 1)",
			},
			"name": map[string]interface{}{
				"type":        "string",
				"description": "Human-readable name for this setup configuration",
			},
			"description": map[string]interface{}{
				"type":        "string",
				"description": "Description of what this setup accomplishes",
			},
			"matrix": map[string]interface{}{
				"type":        "object",
				"description": "OS and architecture constraints",
				"properties": map[string]interface{}{
					"os": map[string]interface{}{
						"type": "array",
						"items": map[string]interface{}{
							"type": "string",
							"enum": []string{"macos", "linux", "windows"},
						},
						"description": "Supported operating systems",
					},
					"arch": map[string]interface{}{
						"type": "array",
						"items": map[string]interface{}{
							"type": "string",
							"enum": []string{"x86_64", "arm64"},
						},
						"description": "Supported architectures",
					},
				},
			},
			"defaults": map[string]interface{}{
				"type":        "object",
				"description": "Default values for steps",
				"properties": map[string]interface{}{
					"timeout": map[string]interface{}{
						"type":        "string",
						"pattern":     "^\\d+[smhd]$",
						"description": "Default timeout for steps (e.g., '15m', '2h')",
					},
					"shell": map[string]interface{}{
						"type":        "string",
						"enum":        []string{"auto", "bash", "sh", "powershell", "cmd"},
						"description": "Shell to use for commands",
					},
					"continue_on_error": map[string]interface{}{
						"type":        "boolean",
						"description": "Whether to continue execution if a step fails",
					},
					"cwd": map[string]interface{}{
						"type":        "string",
						"description": "Default working directory for steps",
					},
				},
			},
			"env": map[string]interface{}{
				"type":        "object",
				"description": "Global environment variables and secrets",
				"properties": map[string]interface{}{
					"secrets": map[string]interface{}{
						"type": "array",
						"items": map[string]interface{}{
							"type": "string",
						},
						"description": "List of secret names to fetch from Initiat",
					},
				},
				"additionalProperties": map[string]interface{}{
					"type": "string",
				},
			},
			"bootstrap": getStepsSchema("Bootstrap phase - guarantee base system prerequisites"),
			"provision": getStepsSchema("Provision phase - install language runtimes and services"),
			"setup":     getStepsSchema("Setup phase - fetch dependencies and configure project"),
			"verify":    getStepsSchema("Verify phase - assert environment works correctly"),
			"post":      getStepsSchema("Post phase - print completion messages"),
		},
		"additionalProperties": false,
	}

	return json.MarshalIndent(schema, "", "  ")
}

func getStepsSchema(description string) map[string]interface{} {
	return map[string]interface{}{
		"type":        "array",
		"description": description,
		"items":       getStepSchema(),
	}
}

func getStepSchema() map[string]interface{} {
	return map[string]interface{}{
		"type":        "object",
		"description": "A single step in a phase",
		"properties": map[string]interface{}{
			"name": map[string]interface{}{
				"type":        "string",
				"description": "Display name for this step",
			},
			"if": map[string]interface{}{
				"type":        "string",
				"description": "Condition expression (e.g., 'os(\"macos\")', 'file_exists(\"package.json\")')",
			},
			"timeout": map[string]interface{}{
				"type":        "string",
				"pattern":     "^\\d+[smhd]$",
				"description": "Step timeout (e.g., '10m', '1h')",
			},
			"cwd": map[string]interface{}{
				"type":        "string",
				"description": "Working directory for this step",
			},
			"env": map[string]interface{}{
				"type":        "object",
				"description": "Environment variables for this step",
				"additionalProperties": map[string]interface{}{
					"type": "string",
				},
			},
			"env_from_secrets": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "string",
				},
				"description": "Secret names to inject as environment variables",
			},
			"optional_secrets": map[string]interface{}{
				"type":        "boolean",
				"description": "Don't fail if secrets are missing",
			},
			"continue_on_error": map[string]interface{}{
				"type":        "boolean",
				"description": "Continue execution even if this step fails",
			},
			"retries": getRetriesSchema(),
		},
		"oneOf": []map[string]interface{}{
			getRunActionSchema(),
			getPrintActionSchema(),
		},
		"additionalProperties": false,
	}
}

func getRetriesSchema() map[string]interface{} {
	return map[string]interface{}{
		"type":        "object",
		"description": "Retry policy for this step",
		"properties": map[string]interface{}{
			"attempts": map[string]interface{}{
				"type":        "integer",
				"minimum":     1,
				"description": "Number of retry attempts",
			},
			"backoff": map[string]interface{}{
				"type":        "string",
				"pattern":     "^\\d+[smhd]$",
				"description": "Delay between retries (e.g., '2s', '1m')",
			},
		},
		"required": []string{"attempts", "backoff"},
	}
}

func getRunActionSchema() map[string]interface{} {
	return map[string]interface{}{
		"properties": map[string]interface{}{
			"run": map[string]interface{}{
				"type":        "string",
				"description": "Shell command to execute",
			},
		},
		"required": []string{"run"},
	}
}

func getPrintActionSchema() map[string]interface{} {
	return map[string]interface{}{
		"properties": map[string]interface{}{
			"print": map[string]interface{}{
				"type":        "string",
				"description": "Message to print",
			},
		},
		"required": []string{"print"},
	}
}

func SaveSchemaToFile(filename string) error {
	schema, err := GenerateJSONSchema()
	if err != nil {
		return fmt.Errorf("failed to generate JSON schema: %w", err)
	}

	if err := os.WriteFile(filename, schema, fileModeReadWrite); err != nil {
		return fmt.Errorf("failed to write schema file: %w", err)
	}

	return nil
}
