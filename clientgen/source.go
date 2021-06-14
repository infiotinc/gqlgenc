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

type TypeTarget struct {
	Type types.Type
	Name string
}

type Type struct {
	Name           string
	Type           types.Type
	UnmarshalTypes map[string]TypeTarget
	RefType        *types.Named
}

func (s *Source) Fragments() error {
	for _, fragment := range s.queryDocument.Fragments {
		name := templates.ToGo(fragment.Name)

		_ = s.sourceGenerator.namedType(name, func() types.Type {
			responseFields := s.sourceGenerator.NewResponseFields(name, &fragment.SelectionSet)

			typ := s.sourceGenerator.genStruct("", name, responseFields)
			return typ
		})
	}

	return nil
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
}

func (s *Source) OperationResponses() ([]*OperationResponse, error) {
	operationResponses := make([]*OperationResponse, 0, len(s.queryDocument.Operations))
	for _, operationResponse := range s.queryDocument.Operations {
		name := getResponseStructName(operationResponse, s.generateConfig)

		opres := &OperationResponse{
			Operation: operationResponse,
			Name:      name,
		}

		namedType := s.sourceGenerator.namedType(name, func() types.Type {
			responseFields := s.sourceGenerator.NewResponseFields(name, &operationResponse.SelectionSet)

			typ := s.sourceGenerator.genStruct("", name, responseFields)
			return typ
		})
		opres.Type = namedType

		operationResponses = append(operationResponses, opres)
	}

	return operationResponses, nil
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
