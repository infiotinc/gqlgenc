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

type SourceGenerator struct {
	cfg    *config.Config
	binder *config.Binder
	client config.PackageConfig
}

func NewSourceGenerator(cfg *config.Config, client config.PackageConfig) *SourceGenerator {
	return &SourceGenerator{
		cfg:    cfg,
		binder: cfg.NewBinder(),
		client: client,
	}
}

func (r *SourceGenerator) NewResponseFields(prefix string, selectionSet ast.SelectionSet) (ResponseFieldList, []*Type) {
	responseFields := make(ResponseFieldList, 0, len(selectionSet))
	genTypes := make([]*Type, 0)
	for _, selection := range selectionSet {
		rf, rfGenTypes := r.NewResponseField(prefix, selection)
		genTypes = append(genTypes, rfGenTypes...)
		responseFields = append(responseFields, rf)
	}

	return responseFields, genTypes
}

func (r *SourceGenerator) namedType(name string, gen func() types.Type) types.Type {
	if r.cfg.Models.Exists(name) {
		model := r.cfg.Models[name].Model[0]
		fmt.Printf("%s is already declared: %v\n", name, model)

		typ, err := r.binder.FindTypeFromName(model)
		if err != nil {
			panic(fmt.Errorf("cannot get type for %v: %w", name, err))
		}

		return typ
	} else {
		r.cfg.Models.Add(
			name,
			fmt.Sprintf("%s.%s", r.client.ImportPath(), name),
		)

		return gen()
	}
}

func (r *SourceGenerator) genStruct(prefix, name string, fieldsResponseFields ResponseFieldList) (types.Type, []*Type) {
	fullName := name
	if prefix != "" {
		fullName = prefix + "_" + fullName
	}

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

	return types.NewNamed(
			types.NewTypeName(0, r.client.Pkg(), fullName, nil),
			types.NewInterfaceType([]*types.Func{}, []types.Type{}),
			nil,
		), []*Type{{
			Name:           fullName,
			Type:           types.NewStruct(vars, tags),
			UnmarshalTypes: unmarshalTypes,
		}}
}

func (r *SourceGenerator) NewResponseField(prefix string, selection ast.Selection) (*ResponseField, []*Type) {
	switch selection := selection.(type) {
	case *ast.Field:
		fieldsResponseFields, genTypes := r.NewResponseFields(prefix, selection.SelectionSet)

		var baseType types.Type
		switch {
		case fieldsResponseFields.IsBasicType():
			baseType = r.Type(selection.Definition.Type.Name())
		case fieldsResponseFields.IsFragment():
			// if a child field is fragment, this field type became fragment.
			baseType = fieldsResponseFields[0].Type
		case fieldsResponseFields.IsStructType():
			typ, rsGenTypes := r.genStruct(prefix, templates.ToGo(selection.Name), fieldsResponseFields)
			genTypes = append(genTypes, rsGenTypes...)
			baseType = typ
		default:
			// ここにきたらバグ
			// here is bug
			panic("not match type")
		}

		// GraphQLの定義がオプショナルのはtypeのポインタ型が返り、配列の定義場合はポインタのスライスの型になって返ってきます
		// return pointer type then optional type or slice pointer then slice type of definition in GraphQL.
		typ := r.binder.CopyModifiersFromAst(selection.Definition.Type, baseType)

		tags := []string{
			fmt.Sprintf(`json:"%s"`, selection.Alias),
		}

		return &ResponseField{
			Name:           selection.Alias,
			Type:           typ,
			Tags:           tags,
			ResponseFields: fieldsResponseFields,
		}, genTypes

	case *ast.FragmentSpread:
		// この構造体はテンプレート側で使われることはなく、ast.FieldでFragment判定するために使用する
		fieldsResponseFields, genTypes := r.NewResponseFields(prefix, selection.Definition.SelectionSet)
		typ, rsGenTypes := r.genStruct(prefix, selection.Name, fieldsResponseFields)
		genTypes = append(genTypes, rsGenTypes...)

		return &ResponseField{
			Name:             selection.Name,
			Type:             typ,
			IsFragmentSpread: true,
			ResponseFields:   fieldsResponseFields,
		}, genTypes

	case *ast.InlineFragment:
		selection.SelectionSet = append(selection.SelectionSet, &ast.Field{
			Name:  "__typename",
			Alias: "__typename",
			Definition: &ast.FieldDefinition{
				Name: "Typename",
				Type: ast.NamedType("__Type", nil),
				Arguments: ast.ArgumentDefinitionList{
					{Name: "name", Type: ast.NonNullNamedType("String", nil)},
				},
			},
		})

		// InlineFragmentは子要素をそのままstructとしてもつので、ここで、構造体の型を作成します
		fieldsResponseFields, genTypes := r.NewResponseFields(prefix, selection.SelectionSet)
		typ, rsGenTypes := r.genStruct(prefix, selection.TypeCondition, fieldsResponseFields)
		genTypes = append(genTypes, rsGenTypes...)

		return &ResponseField{
			Name:             selection.TypeCondition,
			Type:             typ,
			IsInlineFragment: true,
			ResponseFields:   fieldsResponseFields,
		}, genTypes
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
