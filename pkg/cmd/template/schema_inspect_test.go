// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package template_test

import (
	"testing"

	cmdtpl "github.com/k14s/ytt/pkg/cmd/template"
	"github.com/k14s/ytt/pkg/cmd/ui"
	"github.com/k14s/ytt/pkg/files"
	"github.com/stretchr/testify/require"
)

func TestSchemaInspect_exports_an_OpenAPI_doc(t *testing.T) {
	t.Run("for all inferred types with their inferred defaults", func(t *testing.T) {
		opts := cmdtpl.NewOptions()
		opts.DataValuesFlags.InspectSchema = true
		opts.RegularFilesSourceOpts.OutputType.Types = []string{"openapi-v3"}

		schemaYAML := `#@data/values-schema
---
foo:
  int_key: 10
  bool_key: true
  false_key: false
  string_key: some text
  float_key: 9.1
  array_of_scalars:
  - ""
  array_of_maps:
  - foo: ""
    bar: ""
`
		expected := `openapi: 3.0.0
info:
  version: 0.1.0
  title: Schema for data values, generated by ytt
paths: {}
components:
  schemas:
    dataValues:
      type: object
      additionalProperties: false
      properties:
        foo:
          type: object
          additionalProperties: false
          properties:
            int_key:
              type: integer
              default: 10
            bool_key:
              type: boolean
              default: true
            false_key:
              type: boolean
              default: false
            string_key:
              type: string
              default: some text
            float_key:
              type: number
              default: 9.1
              format: float
            array_of_scalars:
              type: array
              items:
                type: string
                default: ""
              default: []
            array_of_maps:
              type: array
              items:
                type: object
                additionalProperties: false
                properties:
                  foo:
                    type: string
                    default: ""
                  bar:
                    type: string
                    default: ""
              default: []
`

		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
		})

		assertSucceedsDocSet(t, filesToProcess, expected, opts)
	})
	t.Run("including explicitly set default values", func(t *testing.T) {
		opts := cmdtpl.NewOptions()
		opts.DataValuesFlags.InspectSchema = true
		opts.RegularFilesSourceOpts.OutputType.Types = []string{"openapi-v3"}

		schemaYAML := `#@data/values-schema
---
foo:
  #@schema/default 10
  int_key: 0
  #@schema/default True
  bool_key: false
  #@schema/default False
  false_key: true
  #@schema/default "some text"
  string_key: ""
  #@schema/default 9.1
  float_key: 0.0
  #@schema/default [1,2,3]
  array_of_scalars:
  - 0
  #@schema/default [{"bar": "thing 1"},{"bar": "thing 2"}, {"bar": "thing 3"}]
  array_of_maps:
  - bar: ""
    ree: "default"
`
		expected := `openapi: 3.0.0
info:
  version: 0.1.0
  title: Schema for data values, generated by ytt
paths: {}
components:
  schemas:
    dataValues:
      type: object
      additionalProperties: false
      properties:
        foo:
          type: object
          additionalProperties: false
          properties:
            int_key:
              type: integer
              default: 10
            bool_key:
              type: boolean
              default: true
            false_key:
              type: boolean
              default: false
            string_key:
              type: string
              default: some text
            float_key:
              type: number
              default: 9.1
              format: float
            array_of_scalars:
              type: array
              items:
                type: integer
                default: 0
              default:
              - 1
              - 2
              - 3
            array_of_maps:
              type: array
              items:
                type: object
                additionalProperties: false
                properties:
                  bar:
                    type: string
                    default: ""
                  ree:
                    type: string
                    default: default
              default:
              - bar: thing 1
                ree: default
              - bar: thing 2
                ree: default
              - bar: thing 3
                ree: default
`

		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
		})

		assertSucceedsDocSet(t, filesToProcess, expected, opts)
	})
	t.Run("including nullable values", func(t *testing.T) {
		opts := cmdtpl.NewOptions()
		opts.DataValuesFlags.InspectSchema = true
		opts.RegularFilesSourceOpts.OutputType.Types = []string{"openapi-v3"}

		schemaYAML := `#@data/values-schema
---
foo:
  #@schema/nullable
  int_key: 0
  #@schema/nullable
  bool_key: false
  #@schema/nullable
  false_key: true
  #@schema/nullable
  string_key: ""
  #@schema/nullable
  float_key: 0.0
  #@schema/nullable
  array_of_scalars:
  - 0
  #@schema/nullable
  array_of_maps:
  -
    #@schema/nullable
    bar: ""
    #@schema/nullable
    ree: ""
`
		expected := `openapi: 3.0.0
info:
  version: 0.1.0
  title: Schema for data values, generated by ytt
paths: {}
components:
  schemas:
    dataValues:
      type: object
      additionalProperties: false
      properties:
        foo:
          type: object
          additionalProperties: false
          properties:
            int_key:
              type: integer
              default: null
              nullable: true
            bool_key:
              type: boolean
              default: null
              nullable: true
            false_key:
              type: boolean
              default: null
              nullable: true
            string_key:
              type: string
              default: null
              nullable: true
            float_key:
              type: number
              default: null
              format: float
              nullable: true
            array_of_scalars:
              type: array
              items:
                type: integer
                default: 0
              default: null
              nullable: true
            array_of_maps:
              type: array
              items:
                type: object
                additionalProperties: false
                properties:
                  bar:
                    type: string
                    default: null
                    nullable: true
                  ree:
                    type: string
                    default: null
                    nullable: true
              default: null
              nullable: true
`

		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
		})

		assertSucceedsDocSet(t, filesToProcess, expected, opts)
	})
	t.Run("including 'any' values", func(t *testing.T) {
		t.Run("on documents", func(t *testing.T) {
			opts := cmdtpl.NewOptions()
			opts.DataValuesFlags.InspectSchema = true
			opts.RegularFilesSourceOpts.OutputType.Types = []string{"openapi-v3"}

			schemaYAML := `#@data/values-schema
#@schema/type any=True
---
foo:
  int_key: 0
  array_of_scalars:
  - ""
  array_of_maps:
  - foo: ""
    bar: ""
`
			expected := `openapi: 3.0.0
info:
  version: 0.1.0
  title: Schema for data values, generated by ytt
paths: {}
components:
  schemas:
    dataValues:
      default:
        foo:
          int_key: 0
          array_of_scalars: []
          array_of_maps: []
`
			filesToProcess := files.NewSortedFiles([]*files.File{
				files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
			})

			assertSucceedsDocSet(t, filesToProcess, expected, opts)
		})
		t.Run("on map items", func(t *testing.T) {
			opts := cmdtpl.NewOptions()
			opts.DataValuesFlags.InspectSchema = true
			opts.RegularFilesSourceOpts.OutputType.Types = []string{"openapi-v3"}

			schemaYAML := `#@data/values-schema
---
#@schema/type any=True
foo:
  int_key: 0
  array_of_scalars:
  - ""
  array_of_maps:
  - foo: ""
    bar: ""
`
			expected := `openapi: 3.0.0
info:
  version: 0.1.0
  title: Schema for data values, generated by ytt
paths: {}
components:
  schemas:
    dataValues:
      type: object
      additionalProperties: false
      properties:
        foo:
          default:
            int_key: 0
            array_of_scalars: []
            array_of_maps: []
`
			filesToProcess := files.NewSortedFiles([]*files.File{
				files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
			})

			assertSucceedsDocSet(t, filesToProcess, expected, opts)
		})
		t.Run("on array items", func(t *testing.T) {
			opts := cmdtpl.NewOptions()
			opts.DataValuesFlags.InspectSchema = true
			opts.RegularFilesSourceOpts.OutputType.Types = []string{"openapi-v3"}

			schemaYAML := `#@data/values-schema
---
foo:
  int_key: 0
  array_of_scalars:
  #@schema/type any=True
  - ""
  array_of_maps:
  - foo: ""
    bar: ""
`
			expected := `openapi: 3.0.0
info:
  version: 0.1.0
  title: Schema for data values, generated by ytt
paths: {}
components:
  schemas:
    dataValues:
      type: object
      additionalProperties: false
      properties:
        foo:
          type: object
          additionalProperties: false
          properties:
            int_key:
              type: integer
              default: 0
            array_of_scalars:
              type: array
              items:
                default: ""
              default: []
            array_of_maps:
              type: array
              items:
                type: object
                additionalProperties: false
                properties:
                  foo:
                    type: string
                    default: ""
                  bar:
                    type: string
                    default: ""
              default: []
`
			filesToProcess := files.NewSortedFiles([]*files.File{
				files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
			})

			assertSucceedsDocSet(t, filesToProcess, expected, opts)
		})
	})
	t.Run("including nullable values with defaults", func(t *testing.T) {
		opts := cmdtpl.NewOptions()
		opts.DataValuesFlags.InspectSchema = true
		opts.RegularFilesSourceOpts.OutputType.Types = []string{"openapi-v3"}

		schemaYAML := `#@data/values-schema
---
foo:
  #@schema/default 10
  #@schema/nullable
  int_key: 0

  #@schema/default True
  #@schema/nullable
  bool_key: false

  #@schema/nullable
  #@schema/default False
  false_key: true

  #@schema/nullable
  #@schema/default "some text"
  string_key: ""

  #@schema/nullable
  #@schema/default 9.1
  float_key: 0.0

  #@schema/nullable
  #@schema/default [1,2,3]
  array_of_scalars:
  - 0

  #@schema/default [{"bar": "thing 1"},{"bar": "thing 2"}, {"bar": "thing 3"}]
  #@schema/nullable
  array_of_maps:
  -
    #@schema/nullable
    bar: ""
    #@schema/nullable
    ree: ""
`
		expected := `openapi: 3.0.0
info:
  version: 0.1.0
  title: Schema for data values, generated by ytt
paths: {}
components:
  schemas:
    dataValues:
      type: object
      additionalProperties: false
      properties:
        foo:
          type: object
          additionalProperties: false
          properties:
            int_key:
              type: integer
              default: 10
              nullable: true
            bool_key:
              type: boolean
              default: true
              nullable: true
            false_key:
              type: boolean
              default: false
              nullable: true
            string_key:
              type: string
              default: some text
              nullable: true
            float_key:
              type: number
              default: 9.1
              format: float
              nullable: true
            array_of_scalars:
              type: array
              items:
                type: integer
                default: 0
              default:
              - 1
              - 2
              - 3
              nullable: true
            array_of_maps:
              type: array
              items:
                type: object
                additionalProperties: false
                properties:
                  bar:
                    type: string
                    default: null
                    nullable: true
                  ree:
                    type: string
                    default: null
                    nullable: true
              default:
              - bar: thing 1
                ree: null
              - bar: thing 2
                ree: null
              - bar: thing 3
                ree: null
              nullable: true
`

		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
		})

		assertSucceedsDocSet(t, filesToProcess, expected, opts)
	})

}

func TestSchemaInspect_errors(t *testing.T) {
	t.Run("when --output is anything other than 'openapi-v3'", func(t *testing.T) {
		opts := cmdtpl.NewOptions()
		opts.DataValuesFlags.InspectSchema = true

		schemaYAML := `#@data/values-schema
---
foo: doesn't matter
`
		expectedErr := "Data values schema export only supported in OpenAPI v3 format; specify format with --output=openapi-v3 flag"

		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
		})

		assertFails(t, filesToProcess, expectedErr, opts)
	})

	t.Run("when --output is set to 'openapi-v3' but not inspecting schema", func(t *testing.T) {
		opts := cmdtpl.NewOptions()
		opts.DataValuesFlags.InspectSchema = false
		opts.RegularFilesSourceOpts.OutputType.Types = []string{"openapi-v3"}

		schemaYAML := `#@data/values-schema
---
foo: doesn't matter
`
		expectedErr := "Output type currently only supported for data values schema (i.e. include --data-values-schema-inspect)"

		filesToProcess := files.NewSortedFiles([]*files.File{
			files.MustNewFileFromSource(files.NewBytesSource("schema.yml", []byte(schemaYAML))),
		})

		assertFails(t, filesToProcess, expectedErr, opts)
	})
}

func assertSucceedsDocSet(t *testing.T, filesToProcess []*files.File, expectedOut string, opts *cmdtpl.Options) {
	t.Helper()
	out := opts.RunWithFiles(cmdtpl.Input{Files: filesToProcess}, ui.NewTTY(false))
	require.NoError(t, out.Err)

	outBytes, err := out.DocSet.AsBytes()
	require.NoError(t, err)

	require.Equal(t, expectedOut, string(outBytes))
}
