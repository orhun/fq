package bson

// https://bsonspec.org/spec.html
// TODO: more types

import (
	"embed"

	"github.com/wader/fq/format"
	"github.com/wader/fq/format/registry"
	"github.com/wader/fq/pkg/decode"
	"github.com/wader/fq/pkg/scalar"
)

//go:embed *.jq
var bsonFS embed.FS

func init() {
	registry.MustRegister(decode.Format{
		Name:        format.BSON,
		Description: "Binary JSON",
		DecodeFn:    decodeBSON,
		Files:       bsonFS,
	})
}

const (
	elementTypeDouble    = 0x01
	elementTypeString    = 0x02
	elementTypeDocument  = 0x03
	elementTypeArray     = 0x04
	elementTypeBinary    = 0x05
	elementTypeUndefined = 0x06
	elementTypeObjectID  = 0x07
	elementTypeBoolean   = 0x08
	elementTypeDatatime  = 0x09
	elementTypeNull      = 0x0a
	elementTypeRegexp    = 0x0b
	elementTypeInt32     = 0x10
	elementTypeInt64     = 0x12
)

var elementTypeMap = scalar.UToScalar{
	elementTypeDouble:    {Sym: "double", Description: "64-bit binary floating point"},
	elementTypeString:    {Sym: "string", Description: "UTF-8 string"},
	elementTypeDocument:  {Sym: "document", Description: "Embedded document"},
	elementTypeArray:     {Sym: "array", Description: "Array"},
	elementTypeBinary:    {Sym: "binary", Description: "Binary data"},
	elementTypeUndefined: {Sym: "Undefined", Description: "Undefined (deprecated)"},
	elementTypeObjectID:  {Sym: "object_id", Description: "ObjectId"},
	elementTypeBoolean:   {Sym: "boolean", Description: "Boolean"},
	elementTypeDatatime:  {Sym: "datatime", Description: "UTC datetime"},
	elementTypeNull:      {Sym: "null", Description: "Null value"},
	elementTypeRegexp:    {Sym: "regexp", Description: "Regular expression"},
	elementTypeInt32:     {Sym: "int32", Description: "32-bit integer"},
	elementTypeInt64:     {Sym: "int64", Description: "64-bit integer"},
}

func decodeBSONDocument(d *decode.D) {
	size := d.FieldU32("size")
	d.LenFn(int64(size-4)*8, func(d *decode.D) {
		d.FieldArray("document", func(d *decode.D) {
			for d.BitsLeft() > 8 {
				d.FieldStruct("element", func(d *decode.D) {
					typ := d.FieldU8("type", elementTypeMap)
					d.FieldUTF8Null("name")
					switch typ {
					case elementTypeDouble:
						d.FieldF64("value")
					case elementTypeString:
						length := d.FieldU32("length")
						d.FieldUTF8NullFixedLen("value", int(length))
					case elementTypeDocument:
						d.FieldStruct("value", decodeBSONDocument)
					case elementTypeBoolean:
						d.FieldU8("value")
					case elementTypeArray:
						d.FieldStruct("value", decodeBSONDocument)
					case elementTypeInt32:
						d.FieldS32("value")
					case elementTypeInt64:
						d.FieldS32("value")
					case elementTypeNull:
					default:
						d.FieldRawLen("value", d.BitsLeft())
					}
				})
			}
		})
		d.FieldU8("terminator")
	})
}

func decodeBSON(d *decode.D, in interface{}) interface{} {
	d.Endian = decode.LittleEndian

	decodeBSONDocument(d)

	return nil
}
