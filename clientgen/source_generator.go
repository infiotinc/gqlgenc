package clientgen

import (
	"fmt"
	"github.com/99designs/gqlgen/codegen/config"
	"github.com/99designs/gqlgen/codegen/templates"
	"github.com/vektah/gqlparser/v2/ast"
	"go/types"
	"strings"
)

type Argument struct {
	Variable string
	Type     types.Type
}

type ResponseField struct {
	Name             string
	IsFragmentSpread bool
	IsInlineFragment bool
	Type             types.Type
	Tags             []string
	ResponseFields   ResponseFieldList
}

type ResponseFieldList []*ResponseField

func (rs ResponseFieldList) IsFragment() bool {
	if len(rs) != 1 {
		return false
	}

	return rs[0].IsInlineFragment || rs[0].IsFragmentSpread
}

func (rs ResponseFieldList) IsBasicType() bool {
	return len(rs) == 0
}

func (rs ResponseFieldList) IsStructType() bool {
	return len(rs) > 0 && !rs.IsFragment()
}

type genType struct {
	name string
	typ  *Type
}

type SourceGenerator struct {
	cfg    *config.Config
	binder *config.Binder
	client config.PackageConfig

	genTypes []genType
}

func (r *SourceGenerator) RegisterGenType(name string, typ *Type) {
	if gt := r.GetGenType(name); gt != nil {
		panic(name + ": gen type already defined")
	}

	r.genTypes = append(r.genTypes, genType{
		name: name,
		typ:  typ,
	})
}

func (r *SourceGenerator) GetGenType(name string) *Type {
	for _, gt := range r.genTypes {
		if gt.name == name {
			return gt.typ
		}
	}

	return nil
}

func (r *SourceGenerator) GenTypes() []*Type {
	typs := make([]*Type, 0)
	for _, gt := range r.genTypes {
		typs = append(typs, gt.typ)
	}

	return typs
}

func NewSourceGenerator(cfg *config.Config, client config.PackageConfig) *SourceGenerator {
	return &SourceGenerator{
		cfg:    cfg,
		binder: cfg.NewBinder(),
		client: client,
	}
}

func (r *SourceGenerator) NewResponseFields(prefix string, selectionSet *ast.SelectionSet) ResponseFieldList {
L:
	for _, field := range *selectionSet {
		switch field.(type) {
		case *ast.InlineFragment:
			*selectionSet = append(ast.SelectionSet{&ast.Field{
				Name:  "__typename",
				Alias: "__typename",
				Definition: &ast.FieldDefinition{
					Name: "Typename",
					Type: ast.NonNullNamedType("String", nil),
					Arguments: ast.ArgumentDefinitionList{
						{Name: "name", Type: ast.NonNullNamedType("String", nil)},
					},
				},
			}}, *selectionSet...)
			break L
		}
	}

	responseFields := make(ResponseFieldList, 0, len(*selectionSet))
	for _, selection := range *selectionSet {
		rf := r.NewResponseField(prefix, selection)
		responseFields = append(responseFields, rf)
	}

	return responseFields
}

func (r *SourceGenerator) namedType(prefix, name string, gen func() types.Type) types.Type {
	fullname := prefixedName(prefix, name)

	if gt := r.GetGenType(fullname); gt != nil {
		return gt.RefType
	}

	if r.cfg.Models.Exists(fullname) {
		model := r.cfg.Models[fullname].Model[0]
		fmt.Printf("%s is already declared: %v\n", fullname, model)

		typ, err := r.binder.FindTypeFromName(model)
		if err != nil {
			panic(fmt.Errorf("cannot get type for %v: %w", fullname, err))
		}

		return typ
	} else {
		r.cfg.Models.Add(
			fullname,
			fmt.Sprintf("%s.%s", r.client.ImportPath(), name),
		)

		return gen()
	}
}

func prefixedName(prefix, name string) string {
	if prefix != "" {
		return prefix + "_" + name
	}

	return name
}

func (r *SourceGenerator) genStruct(prefix, name string, fieldsResponseFields ResponseFieldList) types.Type {
	fullname := prefixedName(prefix, name)

	vars := make([]*types.Var, 0, len(fieldsResponseFields))
	tags := make([]string, 0, len(fieldsResponseFields))
	unmarshalTypes := map[string]TypeTarget{}
	for _, field := range fieldsResponseFields {
		typ := field.Type
		fieldName := templates.ToGo(field.Name)
		if field.IsInlineFragment {
			unmarshalTypes[field.Name] = TypeTarget{
				Type: typ,
				Name: fieldName,
			}
			typ = types.NewPointer(typ)
		}

		vars = append(vars, types.NewVar(0, nil, fieldName, typ))
		tags = append(tags, strings.Join(field.Tags, " "))
	}

	typ := types.NewStruct(vars, tags)

	refType := types.NewNamed(
		types.NewTypeName(0, r.client.Pkg(), fullname, nil),
		nil,
		nil,
	)

	r.RegisterGenType(fullname, &Type{
		Name:           fullname,
		Type:           typ,
		RefType:        refType,
		UnmarshalTypes: unmarshalTypes,
	})

	return refType
}

func (r *SourceGenerator) NewResponseField(prefix string, selection ast.Selection) *ResponseField {
	switch selection := selection.(type) {
	case *ast.Field:
		fieldsResponseFields := r.NewResponseFields(prefix, &selection.SelectionSet)

		var baseType types.Type
		switch {
		case fieldsResponseFields.IsBasicType():
			baseType = r.Type(selection.Definition.Type.Name())
		case fieldsResponseFields.IsFragment():
			// if a child field is fragment, this field type became fragment.
			baseType = fieldsResponseFields[0].Type
		case fieldsResponseFields.IsStructType():
			name := templates.ToGo(selection.Name)
			baseType = r.namedType(prefix, name, func() types.Type {
				return r.genStruct(prefix, name, fieldsResponseFields)
			})
		default:
			// ここにきたらバグ
			// here is bug
			panic("not match type")
		}

		// GraphQLの定義がオプショナルのはtypeのポインタ型が返り、配列の定義場合はポインタのスライスの型になって返ってきます
		// return pointer type then optional type or slice pointer then slice type of definition in GraphQL.
		typ := r.binder.CopyModifiersFromAst(selection.Definition.Type, baseType)

		return &ResponseField{
			Name: selection.Alias,
			Type: typ,
			Tags: []string{
				fmt.Sprintf(`json:"%s"`, selection.Alias),
			},
			ResponseFields: fieldsResponseFields,
		}

	case *ast.FragmentSpread:
		fieldsResponseFields := r.NewResponseFields(prefix, &selection.Definition.SelectionSet)
		typ := r.GetGenType(selection.Definition.Name).RefType

		return &ResponseField{
			Name:             selection.Name,
			Type:             typ,
			IsFragmentSpread: true,
			ResponseFields:   fieldsResponseFields,
		}

	case *ast.InlineFragment:
		// InlineFragmentは子要素をそのままstructとしてもつので、ここで、構造体の型を作成します
		fieldsResponseFields := r.NewResponseFields(prefix, &selection.SelectionSet)
		typ := r.genStruct(prefix, selection.TypeCondition, fieldsResponseFields)

		return &ResponseField{
			Name:             selection.TypeCondition,
			Type:             typ,
			IsInlineFragment: true,
			ResponseFields:   fieldsResponseFields,
			Tags:             []string{`json:"-"`},
		}
	}

	panic("unexpected selection type")
}

func (r *SourceGenerator) OperationArguments(variableDefinitions ast.VariableDefinitionList) []*Argument {
	argumentTypes := make([]*Argument, 0, len(variableDefinitions))
	for _, v := range variableDefinitions {
		argumentTypes = append(argumentTypes, &Argument{
			Variable: v.Variable,
			Type:     r.binder.CopyModifiersFromAst(v.Type, r.Type(v.Type.Name())),
		})
	}

	return argumentTypes
}

// Typeの引数に渡すtypeNameは解析した結果からselectionなどから求めた型の名前を渡さなければいけない
func (r *SourceGenerator) Type(typeName string) types.Type {
	goType, err := r.binder.FindTypeFromName(r.cfg.Models[typeName].Model[0])
	if err != nil {
		// 実装として正しいtypeNameを渡していれば必ず見つかるはずなのでpanic
		panic(fmt.Sprintf("%+v", err))
	}

	return goType
}
