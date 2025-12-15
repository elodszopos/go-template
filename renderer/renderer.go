package renderer

import (
	"bytes"
	"encoding/json"
	"regexp"
	"strconv"
	"text/template"
	"strings"
	"time"
)

type RenderRequest struct {
	Template string          `json:"template"`
	Data     json.RawMessage `json:"data"`
}

type RenderResponse struct {
	Output string `json:"output"`
	Error  string `json:"error,omitempty"`
	Line   int    `json:"line,omitempty"`
	Column *int   `json:"column,omitempty"`
}

type RenderEmailRequest struct {
	SubjectTemplate string          `json:"subjectTemplate"`
	BodyTemplate    string          `json:"bodyTemplate"`
	Data            json.RawMessage `json:"data"`
}

type RenderEmailResponse struct {
	SubjectOutput string `json:"subjectOutput,omitempty"`
	BodyOutput    string `json:"bodyOutput,omitempty"`

	SubjectError  string `json:"subjectError,omitempty"`
	BodyError     string `json:"bodyError,omitempty"`
	SubjectLine   int    `json:"subjectLine,omitempty"`
	BodyLine      int    `json:"bodyLine,omitempty"`
	SubjectColumn *int   `json:"subjectColumn,omitempty"`
	BodyColumn    *int   `json:"bodyColumn,omitempty"`
}

func Render(req RenderRequest) RenderResponse {
	tmpl, err := template.New("template").
		Funcs(TextTemplateFuncMap).
		Parse(req.Template)

	if err != nil {
		return renderErr(err)
	}

	ctx, _, parseErr := buildContext(req.Data)
	if parseErr != nil {
		return RenderResponse{Error: "Data parse error: " + parseErr.Error()}
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, ctx); err != nil {
		return renderErr(err)
	}

	return RenderResponse{Output: buf.String()}
}

func RenderEmail(req RenderEmailRequest) RenderEmailResponse {
	subjectTmpl, err := template.New("subject").
		Funcs(TextTemplateFuncMap).
		Parse(req.SubjectTemplate)

	if err != nil {
		r := renderErr(err)

		return RenderEmailResponse{
			SubjectError:  r.Error,
			SubjectLine:   r.Line,
			SubjectColumn: r.Column,
		}
	}

	bodyTmpl, err := template.New("body").
		Funcs(TextTemplateFuncMap).
		Parse(req.BodyTemplate)

	if err != nil {
		r := renderErr(err)

		return RenderEmailResponse{
			BodyError:  r.Error,
			BodyLine:   r.Line,
			BodyColumn: r.Column,
		}
	}

	ctx, _, parseErr := buildContext(req.Data)
    if parseErr != nil {
    	msg := "Data parse error: " + parseErr.Error()

    	return RenderEmailResponse{
    		SubjectError: msg,
    		BodyError:    msg,
    	}
    }

	var subjectBuf bytes.Buffer
	if err := subjectTmpl.Execute(&subjectBuf, ctx); err != nil {
		r := renderErr(err)

		return RenderEmailResponse{
			SubjectError:  r.Error,
			SubjectLine:   r.Line,
			SubjectColumn: r.Column,
		}
	}

	var bodyBuf bytes.Buffer
	if err := bodyTmpl.Execute(&bodyBuf, ctx); err != nil {
		r := renderErr(err)

		return RenderEmailResponse{
			BodyError:  r.Error,
			BodyLine:   r.Line,
			BodyColumn: r.Column,
		}
	}

	return RenderEmailResponse{
		SubjectOutput: subjectBuf.String(),
		BodyOutput:    bodyBuf.String(),
	}
}

func buildContext(data json.RawMessage) (interface{}, bool, error) {
	{
		var ctx NotificationContext

		dec := json.NewDecoder(strings.NewReader(string(data)))
		dec.DisallowUnknownFields()

		if err := dec.Decode(&ctx); err == nil {
			// Ensure .Now is always available.
			if ctx.Now.IsZero() {
				ctx.Now = time.Now().UTC()
			}

			return &ctx, true, nil
		}
	}

	generic := map[string]interface{}{
		"Now": time.Now().UTC(),
	}

	if len(data) > 0 {
		if err := json.Unmarshal(data, &generic); err != nil {
			return nil, false, err
		}
	}

	if _, ok := generic["Now"]; !ok {
		generic["Now"] = time.Now().UTC()
	}

	return generic, false, nil
}

func renderErr(err error) RenderResponse {
	errMsg := err.Error()
	line, col := extractLineColumn(errMsg)

	var colPtr *int
	if col > 0 {
		colPtr = &col
	}

	return RenderResponse{Error: errMsg, Line: line, Column: colPtr}
}

func extractLineColumn(errMsg string) (int, int) {
	//  - "template: subject:4: ..."
	//  - "template: body:7:3: ..."
	//  - "template: template:7:3: ..."
	re := regexp.MustCompile(`template:\s*[^:]+:(\d+)(?::(\d+))?`)
	matches := re.FindStringSubmatch(errMsg)

	if len(matches) > 1 {
		line, _ := strconv.Atoi(matches[1])
		col := 0

		if len(matches) > 2 && matches[2] != "" {
			col, _ = strconv.Atoi(matches[2])
		}

		return line, col
	}

	return 0, 0
}
