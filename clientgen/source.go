package clientgen

import (
	"bytes"
	"fmt"
	"github.com/99designs/gqlgen/codegen/templates"
	"github.com/infiotinc/gqlgenc/config"
	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/formatter"
	"go/types"
)

type Source struct {
	schema          *ast.Schema
	queryDocument   *ast.QueryDocument
	sourceGenerator *SourceGenerator
	generateConfig  *config.GenerateConfig
}

func NewSource(schema *ast.Schema, queryDocument *ast.QueryDocument, sourceGenerator *SourceGenerator, generateConfig *config.GenerateConfig) *Source {
	return &Source{
		schema:          schema,
		queryDocument:   queryDocument,
		sourceGenerator: sourceGenerator,
		generateConfig:  generateConfig,
	}
}

type Fragment struct {
	Name string
	Type types.Type
}

func (s *Source) Fragments() ([]*Fragment, error) {
	fragments := make([]*Fragment, 0, len(s.queryDocument.Fragments))
	for _, fragment := range s.queryDocument.Fragments {
		responseFields := s.sourceGenerator.NewResponseFields(fragment.SelectionSet)
		name := fragment.Name
		if s.sourceGenerator.cfg.Models.Exists(name) {
			fmt.Printf("%s is already declared: %v\n", name, s.sourceGenerator.cfg.Models[name].Model)
			continue
		}

		fragment := &Fragment{
			Name: name,
			Type: responseFields.StructType(),
		}

		fragments = append(fragments, fragment)
	}

	for _, fragment := range fragments {
		name := fragment.Name
		s.sourceGenerator.cfg.Models.Add(
			name,
			fmt.Sprintf("%s.%s", s.sourceGenerator.client.ImportPath(), templates.ToGo(name)),
		)
	}

	return fragments, nil
}

type Operation struct {
	Name                string
	ResponseType        types.Type
	Operation           string
	OperationType       string
	Args                []*Argument
	VariableDefinitions ast.VariableDefinitionList
}

func NewOperation(operation *OperationResponse, queryDocument *ast.QueryDocument, args []*Argument) *Operation {
	return &Operation{
		Name:                operation.Name,
		OperationType:       string(operation.Operation.Operation),
		ResponseType:        operation.Type,
		Operation:           queryString(queryDocument),
		Args:                args,
		VariableDefinitions: operation.Operation.VariableDefinitions,
	}
}

func (s *Source) Operations(queryDocuments []*ast.QueryDocument, operationResponses []*OperationResponse) []*Operation {
	operations := make([]*Operation, 0, len(s.queryDocument.Operations))

	queryDocumentsMap := queryDocumentMapByOperationName(queryDocuments)
	operationArgsMap := s.operationArgsMapByOperationName()
	for _, operation := range operationResponses {
		queryDocument := queryDocumentsMap[operation.Name]
		args := operationArgsMap[operation.Name]
		operations = append(operations, NewOperation(operation, queryDocument, args))
	}

	return operations
}

func (s *Source) operationArgsMapByOperationName() map[string][]*Argument {
	operationArgsMap := make(map[string][]*Argument)
	for _, operation := range s.queryDocument.Operations {
		operationArgsMap[operation.Name] = s.sourceGenerator.OperationArguments(operation.VariableDefinitions)
	}

	return operationArgsMap
}

func queryDocumentMapByOperationName(queryDocuments []*ast.QueryDocument) map[string]*ast.QueryDocument {
	queryDocumentMap := make(map[string]*ast.QueryDocument)
	for _, queryDocument := range queryDocuments {
		operation := queryDocument.Operations[0]
		queryDocumentMap[operation.Name] = queryDocument
	}

	return queryDocumentMap
}

func queryString(queryDocument *ast.QueryDocument) string {
	var buf bytes.Buffer
	astFormatter := formatter.NewFormatter(&buf)
	astFormatter.FormatQueryDocument(queryDocument)

	return buf.String()
}

type OperationResponse struct {
	Operation *ast.OperationDefinition
	Name      string
	Type      types.Type
	Gen       bool
}

func (s *Source) OperationResponses() ([]*OperationResponse, error) {
	operationResponses := make([]*OperationResponse, 0, len(s.queryDocument.Operations))
	for _, operationResponse := range s.queryDocument.Operations {
		responseFields := s.sourceGenerator.NewResponseFields(operationResponse.SelectionSet)
		name := getResponseStructName(operationResponse, s.generateConfig)

		var typ types.Type
		var gen bool
		if s.sourceGenerator.cfg.Models.Exists(name) {
			fmt.Printf("%s is already declared: %v\n", name, s.sourceGenerator.cfg.Models[name].Model)
			gen = false
			typ = s.sourceGenerator.Type(name)
		} else {
			s.sourceGenerator.cfg.Models.Add(
				name,
				fmt.Sprintf("%s.%s", s.sourceGenerator.client.ImportPath(), templates.ToGo(name)),
			)

			gen = true
			typ = responseFields.StructType()
		}

		operationResponses = append(operationResponses, &OperationResponse{
			Operation: operationResponse,
			Name:      name,
			Type:      typ,
			Gen:       gen,
		})
	}

	return operationResponses, nil
}

type Query struct {
	Name string
	Type types.Type
}

func (s *Source) Query() (*Query, error) {
	fields, err := s.sourceGenerator.NewResponseFieldsByDefinition(s.schema.Query)
	if err != nil {
		return nil, fmt.Errorf("generate failed for query struct type : %w", err)
	}

	s.sourceGenerator.cfg.Models.Add(
		s.schema.Query.Name,
		fmt.Sprintf("%s.%s", s.sourceGenerator.client.Pkg(), templates.ToGo(s.schema.Query.Name)),
	)

	return &Query{
		Name: s.schema.Query.Name,
		Type: fields.StructType(),
	}, nil
}

type Mutation struct {
	Name string
	Type types.Type
}

func (s *Source) Mutation() (*Mutation, error) {
	fields, err := s.sourceGenerator.NewResponseFieldsByDefinition(s.schema.Mutation)
	if err != nil {
		return nil, fmt.Errorf("generate failed for mutation struct type : %w", err)
	}

	s.sourceGenerator.cfg.Models.Add(
		s.schema.Mutation.Name,
		fmt.Sprintf("%s.%s", s.sourceGenerator.client.Pkg(), templates.ToGo(s.schema.Mutation.Name)),
	)

	return &Mutation{
		Name: s.schema.Mutation.Name,
		Type: fields.StructType(),
	}, nil
}

type Subscription struct {
	Name string
	Type types.Type
}

func (s *Source) Subscription() (*Subscription, error) {
	fields, err := s.sourceGenerator.NewResponseFieldsByDefinition(s.schema.Subscription)
	if err != nil {
		return nil, fmt.Errorf("generate failed for subscription struct type : %w", err)
	}

	s.sourceGenerator.cfg.Models.Add(
		s.schema.Subscription.Name,
		fmt.Sprintf("%s.%s", s.sourceGenerator.client.Pkg(), templates.ToGo(s.schema.Subscription.Name)),
	)

	return &Subscription{
		Name: s.schema.Subscription.Name,
		Type: fields.StructType(),
	}, nil
}

func getResponseStructName(operation *ast.OperationDefinition, generateConfig *config.GenerateConfig) string {
	name := operation.Name
	if generateConfig != nil {
		if generateConfig.Prefix != nil {
			if operation.Operation == ast.Subscription {
				name = fmt.Sprintf("%s%s", generateConfig.Prefix.Subscription, name)
			}

			if operation.Operation == ast.Mutation {
				name = fmt.Sprintf("%s%s", generateConfig.Prefix.Mutation, name)
			}

			if operation.Operation == ast.Query {
				name = fmt.Sprintf("%s%s", generateConfig.Prefix.Query, name)
			}
		}

		if generateConfig.Suffix != nil {
			if operation.Operation == ast.Subscription {
				name = fmt.Sprintf("%s%s", name, generateConfig.Suffix.Subscription)
			}

			if operation.Operation == ast.Mutation {
				name = fmt.Sprintf("%s%s", name, generateConfig.Suffix.Mutation)
			}

			if operation.Operation == ast.Query {
				name = fmt.Sprintf("%s%s", name, generateConfig.Suffix.Query)
			}
		}
	}

	return name
}
