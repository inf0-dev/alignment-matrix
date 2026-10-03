package renderer

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"strings"

	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/css"
	"github.com/tdewolff/minify/v2/js"

	v1 "github.com/inf0-dev/alma/api/v1"
)

//go:embed templates/page.html.tmpl
var pageTmpl string

//go:embed templates/style.css
var rawCSS string

//go:embed templates/script.js
var rawJS string

//go:embed templates/*
var _ embed.FS

var (
	minCSS string
	minJS  string
)

func init() {
	m := minify.New()
	m.AddFunc("text/css", css.Minify)
	m.AddFunc("application/javascript", js.Minify)

	var err error
	minCSS, err = m.String("text/css", rawCSS)
	if err != nil {
		minCSS = rawCSS // fallback to unminified
	}
	minJS, err = m.String("application/javascript", rawJS)
	if err != nil {
		minJS = rawJS
	}
}

// ReqGridCell represents one square in the commit-graph style requirement grid.
type ReqGridCell struct {
	ID     string
	Desc   string
	Type   string // "hard" or "soft"
	Status string // "met", "partial", or empty
}

var funcMap = template.FuncMap{
	"joinStrings":     strings.Join,
	"requirementDesc": requirementDesc,
	"joinFailedReqs":  joinFailedReqs,
	"schemaItem":      schemaItem,
	"schemaOption":    schemaOption,
	"reqGridCells":    reqGridCells,
	"derefOr":         derefOr,
	"derefInt":        func(p *int) int { return *p },
	"string":          func(k v1.Kind) string { return string(k) },
}

var htmlTemplate = template.Must(
	template.New("page").Funcs(funcMap).Parse(pageTmpl),
)

type pageData struct {
	v1.Record
	CSS              template.CSS
	JS               template.JS
	RequirementsJSON template.JS
	ItemsJSON        template.JS
	SchemaJSON       template.JS
}

// HTML renders a Record as a self-contained HTML page.
func HTML(r *v1.Record) (string, error) {
	reqJSON, err := json.Marshal(r.Requirements)
	if err != nil {
		return "", fmt.Errorf("failed to marshal requirements: %w", err)
	}

	itemsJSON, err := json.Marshal(r.Items)
	if err != nil {
		return "", fmt.Errorf("failed to marshal items: %w", err)
	}

	schemaJSON, err := json.Marshal(r.Model.Schema)
	if err != nil {
		return "", fmt.Errorf("failed to marshal schema: %w", err)
	}

	data := pageData{
		Record:           *r,
		CSS:              template.CSS(minCSS),
		JS:               template.JS(minJS),
		RequirementsJSON: template.JS(reqJSON),
		ItemsJSON:        template.JS(itemsJSON),
		SchemaJSON:       template.JS(schemaJSON),
	}

	var buf bytes.Buffer
	if err := htmlTemplate.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to render HTML: %w", err)
	}
	return buf.String(), nil
}

func requirementDesc(doc v1.Document, id string) string {
	for _, req := range doc.Schema.Requirements {
		if req.ID == id {
			return req.Description
		}
	}
	return id
}

func joinFailedReqs(doc v1.Document, ids []string) string {
	descs := make([]string, len(ids))
	for i, id := range ids {
		descs[i] = requirementDesc(doc, id)
	}
	return strings.Join(descs, ", ")
}

func schemaItem(doc v1.Document, id string) v1.Item {
	for _, item := range doc.Schema.Items {
		if item.ID == id {
			return item
		}
	}
	return v1.Item{}
}

func schemaOption(doc v1.Document, id string) v1.DesignOption {
	for _, opt := range doc.Schema.DesignOptions {
		if opt.ID == id {
			return opt
		}
	}
	return v1.DesignOption{}
}

func reqGridCells(doc v1.Document, optionID string) []ReqGridCell {
	opt := schemaOption(doc, optionID)
	var cells []ReqGridCell
	for _, req := range doc.Schema.Requirements {
		cell := ReqGridCell{
			ID:   req.ID,
			Desc: req.Description,
		}
		if req.IsHard {
			cell.Type = "hard"
		} else {
			cell.Type = "soft"
		}
		if status, ok := opt.RequirementsMet[req.ID]; ok {
			if status.Met {
				cell.Status = "met"
			} else if status.Partial {
				cell.Status = "partial"
			}
		}
		cells = append(cells, cell)
	}
	return cells
}

// derefOr finds the current answer for an item, returns "" if unanswered.
func derefOr(items []v1.RecordItem, itemID string) string {
	for _, item := range items {
		if item.ID == itemID && item.Answer != nil {
			return *item.Answer
		}
	}
	return ""
}
