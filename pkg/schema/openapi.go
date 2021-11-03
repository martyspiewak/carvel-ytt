// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package schema

import (
	"fmt"

	"github.com/k14s/ytt/pkg/yamlmeta"
)

// OpenAPIDocument holds the document type used for creating an OpenAPI document
type OpenAPIDocument struct {
	docType *DocumentType
}

// NewOpenAPIDocument creates an instance of an OpenAPIDocument based on the given DocumentType
func NewOpenAPIDocument(docType *DocumentType) *OpenAPIDocument {
	return &OpenAPIDocument{docType}
}

// AsDocument generates a new AST of this OpenAPI v3.0.x document, populating the `schemas:` section with the
// type information contained in `docType`.
func (o *OpenAPIDocument) AsDocument() *yamlmeta.Document {
	openAPIProperties := o.calculateProperties(o.docType)

	return &yamlmeta.Document{Value: &yamlmeta.Map{Items: []*yamlmeta.MapItem{
		{Key: "openapi", Value: "3.0.0"},
		{Key: "info", Value: &yamlmeta.Map{Items: []*yamlmeta.MapItem{
			{Key: "version", Value: "0.1.0"},
			{Key: "title", Value: "Schema for data values, generated by ytt"},
		}}},
		{Key: "paths", Value: &yamlmeta.Map{}},
		{Key: "components", Value: &yamlmeta.Map{Items: []*yamlmeta.MapItem{
			{Key: "schemas", Value: &yamlmeta.Map{Items: []*yamlmeta.MapItem{
				{Key: "dataValues", Value: openAPIProperties},
			}}},
		}}},
	}}}
}

func (o *OpenAPIDocument) calculateProperties(schemaVal interface{}) *yamlmeta.Map {
	switch typedValue := schemaVal.(type) {
	case *DocumentType:
		return o.calculateProperties(typedValue.GetValueType())
	case *MapType:
		var properties []*yamlmeta.MapItem
		for _, i := range typedValue.Items {
			mi := yamlmeta.MapItem{Key: i.Key, Value: o.calculateProperties(i.GetValueType())}
			properties = append(properties, &mi)
		}
		property := yamlmeta.Map{Items: []*yamlmeta.MapItem{
			{Key: "type", Value: "object"},
			{Key: "additionalProperties", Value: false},
			{Key: "properties", Value: &yamlmeta.Map{Items: properties}},
		}}
		return &property
	case *ArrayType:
		valueType := typedValue.GetValueType().(*ArrayItemType)
		properties := o.calculateProperties(valueType.GetValueType())
		property := yamlmeta.Map{Items: []*yamlmeta.MapItem{
			{Key: "type", Value: "array"},
			{Key: "items", Value: properties},
			{Key: "default", Value: typedValue.GetDefaultValue()},
		}}
		return &property
	case *ScalarType:
		typeString := o.openAPITypeFor(typedValue)
		defaultVal := typedValue.GetDefaultValue()
		property := yamlmeta.Map{Items: []*yamlmeta.MapItem{
			{Key: "type", Value: typeString},
			{Key: "default", Value: defaultVal},
		}}
		if typedValue.String() == "float" {
			property.Items = append(property.Items, &yamlmeta.MapItem{Key: "format", Value: "float"})
		}
		return &property
	case *NullType:
		properties := o.calculateProperties(typedValue.GetValueType())
		properties.Items = append(properties.Items, &yamlmeta.MapItem{Key: "nullable", Value: true})
		return properties
	case *AnyType:
		return &yamlmeta.Map{Items: []*yamlmeta.MapItem{
			{Key: "default", Value: typedValue.GetDefaultValue()},
		}}
	default:
		panic(fmt.Sprintf("Unrecognized type %T", schemaVal))
	}
}

func (o *OpenAPIDocument) openAPITypeFor(astType *ScalarType) string {
	switch astType.ValueType.(type) {
	case string:
		return "string"
	case float64:
		return "number"
	case int:
		return "integer"
	case bool:
		return "boolean"
	default:
		panic(fmt.Sprintf("Unrecognized type: %T", astType.ValueType))
	}
}
