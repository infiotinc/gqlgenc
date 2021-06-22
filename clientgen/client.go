package clientgen

import (
	"fmt"
	"github.com/99designs/gqlgen/codegen/config"
	"github.com/99designs/gqlgen/plugin"
	"github.com/99designs/gqlgen/plugin/modelgen"
	gqlgencConfig "github.com/infiotinc/gqlgenc/config"
	"github.com/vektah/gqlparser/v2/ast"
)

var _ plugin.ConfigMutator = &Plugin{}

type Plugin struct {
	queryFilePaths []string
	Client         config.PackageConfig
	GenerateConfig *gqlgencConfig.GenerateConfig
	ExtraTypes     []string
}

func New(queryFilePaths []string, client config.PackageConfig, generateConfig *gqlgencConfig.GenerateConfig, extraTypes []string) *Plugin {
	return &Plugin{
		queryFilePaths: queryFilePaths,
		Client:         client,
		GenerateConfig: generateConfig,
		ExtraTypes:     extraTypes,
	}
}

func (p *Plugin) Name() string {
	return "clientgen"
}

func (p *Plugin) ShouldGenType(def *ast.Definition) bool {
	if def.BuiltIn {
		return true
	}

	if def.IsInputType() {
		return true
	}

	for _, t := range p.ExtraTypes {
		if t == def.Name {
			return true
		}
	}

	return false
}

// ModelGenMutateConfig Only uses modelgen for specific types
func (p *Plugin) ModelGenMutateConfig(cfg *config.Config) error {
	schema := &ast.Schema{
		Types: map[string]*ast.Definition{},
	}
	for name, def := range cfg.Schema.Types {
		if p.ShouldGenType(def) {
			schema.Types[name] = def
		}
	}

	return modelgen.New().(*modelgen.Plugin).MutateConfig(&config.Config{
		Model:                    cfg.Model,
		Federation:               cfg.Federation,
		Resolver:                 cfg.Resolver,
		Models:                   cfg.Models,
		StructTag:                cfg.StructTag,
		Directives:               cfg.Directives,
		OmitSliceElementPointers: cfg.OmitSliceElementPointers,
		SkipValidation:           cfg.SkipValidation,
		Sources:                  cfg.Sources,
		Packages:                 cfg.Packages,
		Schema:                   schema,
		Federated:                cfg.Federated,
	})
}

func (p *Plugin) MutateConfig(cfg *config.Config) error {
	err := p.ModelGenMutateConfig(cfg)
	if err != nil {
		return fmt.Errorf("modelgen: %w", err)
	}

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

	err = source.Fragments()
	if err != nil {
		return fmt.Errorf("generating fragment failed: %w", err)
	}

	operationResponses, err := source.OperationResponses()
	if err != nil {
		return fmt.Errorf("generating operation response failed: %w", err)
	}

	operations := source.Operations(queryDocuments, operationResponses)

	types := sourceGenerator.GenTypes()

	generateClient := p.GenerateConfig.ShouldGenerateClient()
	if err := RenderTemplate(cfg, types, operations, operationResponses, generateClient, p.Client); err != nil {
		return fmt.Errorf("template failed: %w", err)
	}

	return nil
}
