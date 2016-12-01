package golang

import (
	"text/template"

	"github.com/eyrie-io/gencode/schema"
)

var (
	IntTemps *template.Template
)

func init() {
	IntTemps = template.New("IntTemps").Funcs(template.FuncMap{
		"Bytes": func(bits int) int {
			return bits / 8
		},
		"BitRange": func(bits int) []int {
			return []int{0, 8, 16, 24, 32, 40, 48, 56, 64}[0:(bits / 8)]
		},
	})

	template.Must(IntTemps.New("marshal").Parse(`
	{
		{{if .VarInt }}
		t := uint{{.Bits}}({{.Target}})
		{{if .Signed}}
		t <<= 1
   		if {{.Target}} < 0 {
   			t = ^t
   		}
		{{end}}
		for t >= 0x80 {
			buf[i] = byte(t) | 0x80
			t >>= 7
			i++
		}
		buf[i] = byte(t)
		i++
		{{else}}
		{{if .W.Unsafe}}
		*(*{{if not .Signed}}u{{end}}int{{.Bits}})(unsafe.Pointer(&buf[i])) = {{.Target}}
		{{else}}
		{{range BitRange .Bits}}
		buf[i + {{Bytes .}}] = byte({{$.Target}} >> {{.}})
		{{end}}
		{{end}}
		i += {{.Bits}}/8
		{{end}}
	}`))
	template.Must(IntTemps.New("unmarshal").Parse(`
	{
		{{if .VarInt}}
		bs := uint8(7)
		t := uint{{.Bits}}(buf[i] & 0x7F)
		for buf[i] & 0x80 == 0x80 {
			i++
			t |= uint{{.Bits}}(buf[i]&0x7F) << bs
			bs += 7
		}
		i++
		{{if .Signed}}
		{{.Target}} = int{{.Bits}}(t >> 1)
		if t&1 != 0 {
			{{.Target}} = ^{{.Target}}
		}
		{{else}}
		{{.Target}} = t
		{{end}}
		{{else}}
		{{if .W.Unsafe}}
		{{.Target}} = *(*{{if not .Signed}}u{{end}}int{{.Bits}})(unsafe.Pointer(&buf[i]))
		{{else}}
		{{$.Target}} = 0{{range BitRange .Bits}} | ({{if not $.Signed}}u{{end}}int{{$.Bits}}(buf[i + {{Bytes .}}]) << {{.}}){{end}}
		{{end}}
		i += {{.Bits}}/8
		{{end}}
	}`))
	template.Must(IntTemps.New("field").Parse(`{{if not .Signed}}u{{end}}int{{.Bits}}`))
	template.Must(IntTemps.New("size").Parse(`
	{
		{{if .VarInt}}
		{{if .Signed}}
		t := uint{{.Bits}}({{.Target}})
		t <<= 1
		if {{.Target}} < 0 {
			t = ^t
		}
		for t >= 0x80 {
			t >>= 7
			s++
		}
		s++
		{{else}}
		t := {{.Target}}
		for t >= 0x80 {
			t >>= 7
			s++
		}
		s++
		{{end}}
		{{else}}
		s += {{.Bits}}/8
		{{end}}
	}`))
}

type IntTemp struct {
	*schema.IntType
	W      *Walker
	Target string
}

func (w *Walker) WalkIntDef(it *schema.IntType) (parts *StringBuilder, err error) {
	parts = &StringBuilder{}
	err = parts.AddTemplate(IntTemps, "field", it)
	return
}

func (w *Walker) WalkIntSize(it *schema.IntType, target string) (parts *StringBuilder, err error) {
	parts = &StringBuilder{}
	err = parts.AddTemplate(IntTemps, "size", IntTemp{it, w, target})
	return
}

func (w *Walker) WalkIntMarshal(it *schema.IntType, target string) (parts *StringBuilder, err error) {
	parts = &StringBuilder{}
	err = parts.AddTemplate(IntTemps, "marshal", IntTemp{it, w, target})
	return
}

func (w *Walker) WalkIntUnmarshal(it *schema.IntType, target string) (parts *StringBuilder, err error) {
	parts = &StringBuilder{}
	err = parts.AddTemplate(IntTemps, "unmarshal", IntTemp{it, w, target})
	return
}
