package main

import (
	"encoding/json"
	"syscall/js"

	"go-template/renderer"
)

func render(this js.Value, args []js.Value) any {
	if len(args) < 2 {
		resp := renderer.RenderResponse{Error: "Expected (template: string, dataJson: string)"}
		b, _ := json.Marshal(resp)
		return string(b)
	}

	templateText := args[0].String()
	dataJSON := args[1].String()

	req := renderer.RenderRequest{
		Template: templateText,
		Data:     json.RawMessage(dataJSON),
	}

	resp := renderer.Render(req)
	b, _ := json.Marshal(resp)

	return string(b)
}

func renderEmail(this js.Value, args []js.Value) any {
	if len(args) < 3 {
		resp := renderer.RenderEmailResponse{
			SubjectError: "Expected (subjectTemplate: string, bodyTemplate: string, dataJson: string)",
			BodyError:    "Expected (subjectTemplate: string, bodyTemplate: string, dataJson: string)",
		}
		b, _ := json.Marshal(resp)
		return string(b)
	}

	subjectTemplate := args[0].String()
	bodyTemplate := args[1].String()
	dataJSON := args[2].String()

	req := renderer.RenderEmailRequest{
		SubjectTemplate: subjectTemplate,
		BodyTemplate:    bodyTemplate,
		Data:            json.RawMessage(dataJSON),
	}

	resp := renderer.RenderEmail(req)
	b, _ := json.Marshal(resp)

	return string(b)
}

func main() {
	js.Global().Set("goTemplateRender", js.FuncOf(render))
	js.Global().Set("goTemplateRenderEmail", js.FuncOf(renderEmail))
	select {}
}
