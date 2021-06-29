package clientgen

import (
	"fmt"
	"github.com/99designs/gqlgen/codegen/config"
	"github.com/99designs/gqlgen/plugin"
	gqlgencConfig "github.com/infiotinc/gqlgenc/config"
	"go/types"
)

var _ plugin.ConfigMutator = &Plugin{}

type Plugin struct {
	queryFilePaths []string
	Client         config.PackageConfig
	GenerateConfig *gqlgencConfig.GenerateConfig
	ExtraTypes     []string
}

func New(queryFilePaths []string, client gqlgencConfig.Client, generateConfig *gqlgencConfig.GenerateConfig) *Plugin {
	return &Plugin{
		queryFilePaths: queryFilePaths,
		Client:         client.PackageConfig,
		GenerateConfig: generateConfig,
		ExtraTypes:     client.ExtraTypes,
	}
}

func (p *Plugin) Name() string {
	return "clientgen"
}

func (p *Plugin) MutateConfig(cfg *config.Config) error {
	querySources, err := LoadQuerySources(p.queryFilePaths)
	if err != nil {
		return fmt.Errorf("load query sources failed: %w", err)
	}

	// 1. 全体のqueryDocumentを1度にparse
	// 1. Parse document from source of query
	queryDocument, err := ParseQueryDocuments(cfg.Schema, querySources)
	if err != nil {
		return fmt.Errorf(": %w", err)
	}

	// 2. OperationごとのqueryDocumentを作成
	// 2. Separate documents for each operation
	queryDocuments, err := QueryDocumentsByOperations(cfg.Schema, queryDocument.Operations)
	if err != nil {
		return fmt.Errorf("parse query document failed: %w", err)
	}

	// 3. テンプレートと情報ソースを元にコード生成
	// 3. Generate code from template and document source
	sourceGenerator := NewSourceGenerator(cfg, p.Client)
	source := NewSource(cfg.Schema, queryDocument, sourceGenerator, p.GenerateConfig)

	for _, t := range p.ExtraTypes {
		def := cfg.Schema.Types[t]

		if def == nil {
			panic("type " + t + " does not exist in schema")
		}

		_ = sourceGenerator.namedType("", def.Name, func() types.Type {
			return sourceGenerator.GenFromDefinition(def.Name, def)
		})
	}

	err = source.Fragments()
	if err != nil {
		return fmt.Errorf("generating fragment failed: %w", err)
	}

	operationResponses, err := source.OperationResponses()
	if err != nil {
		return fmt.Errorf("generating operation response failed: %w", err)
	}

	operations := source.Operations(queryDocuments, operationResponses)

	genTypes := sourceGenerator.GenTypes()

	generateClient := p.GenerateConfig.ShouldGenerateClient()
	if err := RenderTemplate(cfg, genTypes, operations, generateClient, p.Client); err != nil {
		return fmt.Errorf("template failed: %w", err)
	}

	return nil
}
