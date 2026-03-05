package codeanalysis

import (
	"fmt"

	"github.com/InitiatDev/initiat-cli/internal/codeanalysis/engine"
	elixirlang "github.com/InitiatDev/initiat-cli/internal/codeanalysis/lang/elixir"
	golanglang "github.com/InitiatDev/initiat-cli/internal/codeanalysis/lang/go"
	jslang "github.com/InitiatDev/initiat-cli/internal/codeanalysis/lang/javascript"
	pylang "github.com/InitiatDev/initiat-cli/internal/codeanalysis/lang/python"
	rubylang "github.com/InitiatDev/initiat-cli/internal/codeanalysis/lang/ruby"
	tslang "github.com/InitiatDev/initiat-cli/internal/codeanalysis/lang/typescript"
)

func DefaultRegistry() (*engine.Registry, error) {
	r := engine.NewRegistry()
	langs := []engine.Language{
		golanglang.New(),
		jslang.New(),
		tslang.New(),
		pylang.New(),
		rubylang.New(),
		elixirlang.New(),
	}
	for _, lang := range langs {
		if err := r.Register(lang); err != nil {
			return nil, fmt.Errorf("register %s: %w", lang.ID(), err)
		}
	}
	return r, nil
}
