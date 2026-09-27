package admin

import (
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"strings"

	"github.com/wolfandwings/api/internal/model"
)

//go:embed views/*.html
var viewsFS embed.FS

type Templates struct {
	pages map[string]*template.Template
}

func parseTemplates() *Templates {
	funcs := template.FuncMap{
		"statusCls":    statusCls,
		"statusLabel":  statusLabel,
		"variantsText": variantsText,
		"preorderCls":  preorderCls,
		"kpk2rub":      kpk2rub,
		"kpk2text":     kpk2text,
		"imagesText":   imagesText,
	}

	layoutData, _ := viewsFS.ReadFile("views/layout.html")
	base := template.Must(template.New("layout").Funcs(funcs).Parse(string(layoutData)))

	pages := map[string]*template.Template{}
	entries, _ := fs.ReadDir(viewsFS, "views")
	for _, e := range entries {
		name := e.Name()
		if name == "layout.html" {
			continue
		}
		data, _ := viewsFS.ReadFile("views/" + name)
		t, err := base.Clone()
		if err != nil {
			panic(err)
		}
		pages[strings.TrimSuffix(name, ".html")] = template.Must(t.Parse(string(data)))
	}
	return &Templates{pages: pages}
}

// WritePage рендерит layout целиком в ResponseWriter.
func (t *Templates) WritePage(w interface{ Write([]byte) (int, error) }, name string, data any) error {
	tpl, ok := t.pages[name]
	if !ok {
		return fs.ErrNotExist
	}
	return tpl.ExecuteTemplate(w, "layout", data)
}

func statusCls(s string) string {
	switch s {
	case "in_stock":
		return "b-ok"
	case "preorder":
		return "b-pre"
	case "archived", "draft":
		return "b-arc"
	}
	return ""
}

func statusLabel(s string) string {
	switch s {
	case "in_stock":
		return "в наличии"
	case "preorder":
		return "предзаказ"
	case "low":
		return "мало"
	case "archived":
		return "архив"
	case "draft":
		return "черновик"
	}
	return s
}

func preorderCls(s string) string {
	switch s {
	case "new":
		return "b-pre"
	case "confirmed":
		return "b-ok"
	case "cancelled":
		return "b-arc"
	}
	return ""
}

func variantsText(vs []model.Variant) string {
	parts := make([]string, 0, len(vs))
	for _, v := range vs {
		parts = append(parts, v.Name+" | "+itoa(v.PriceDelta))
	}
	return strings.Join(parts, "\n")
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	sign := ""
	if n < 0 {
		sign = "-"
		n = -n
	}
	digits := []byte{}
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	return sign + string(digits)
}

// kpk2rub конвертирует копейки в рубли (для input value).
func kpk2rub(kopecks int) string {
	rub := float64(kopecks) / 100.0
	return fmt.Sprintf("%.2f", rub)
}

// kpk2text форматирует копейки для отображения.
func kpk2text(kopecks int) string {
	rub := float64(kopecks) / 100.0
	return fmt.Sprintf("%.0f ₽", rub)
}

// imagesText форматирует фото для textarea.
func imagesText(ims []model.Image) string {
	parts := make([]string, 0, len(ims))
	for _, im := range ims {
		prim := "0"
		if im.IsPrimary {
			prim = "1"
		}
		parts = append(parts, im.URL+" | "+im.Alt+" | "+prim)
	}
	return strings.Join(parts, "\n")
}
