// Package docx builds native editable OOXML documents from the same normalized CV model.
package docx

import (
	"archive/zip"
	"bytes"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"strings"

	"github.com/troubleman96/tengeneza-cv-desktop/internal/render/model"
)

type Renderer struct{}

const photoEMU = 900000

func (Renderer) Render(doc model.Document) ([]byte, error) {
	var out bytes.Buffer
	archive := zip.NewWriter(&out)
	ext, data := photoData(doc)
	files := map[string]string{"[Content_Types].xml": contentTypes, "_rels/.rels": rootRels, "word/_rels/document.xml.rels": documentRels, "word/styles.xml": styles, "word/numbering.xml": numbering, "word/document.xml": documentXML(doc, ext != "")}
	if ext != "" {
		files["[Content_Types].xml"] = strings.Replace(contentTypes, "</Types>", fmt.Sprintf(`<Default Extension="%s" ContentType="image/%s"/></Types>`, ext, ext), 1)
		files["word/_rels/document.xml.rels"] = strings.Replace(documentRels, "</Relationships>", `<Relationship Id="rId3" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/image" Target="media/profile.`+ext+`"/></Relationships>`, 1)
	}
	for name, content := range files {
		writer, err := archive.Create(name)
		if err != nil {
			return nil, err
		}
		if _, err := writer.Write([]byte(content)); err != nil {
			return nil, err
		}
	}
	if ext != "" {
		writer, err := archive.Create("word/media/profile." + ext)
		if err != nil {
			return nil, err
		}
		if _, err := writer.Write(data); err != nil {
			return nil, err
		}
	}
	if err := archive.Close(); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}
func photoData(doc model.Document) (string, []byte) {
	source := string(doc.Personal.PhotoURL)
	const marker = ";base64,"
	if !strings.HasPrefix(source, "data:image/") || !strings.Contains(source, marker) {
		return "", nil
	}
	mime := strings.TrimPrefix(source[:strings.Index(source, marker)], "data:")
	raw, err := base64.StdEncoding.DecodeString(source[strings.Index(source, marker)+len(marker):])
	if err != nil || len(raw) == 0 {
		return "", nil
	}
	switch mime {
	case "image/jpeg":
		return "jpg", raw
	case "image/png":
		return "png", raw
	default:
		return "", nil
	}
}
func photoParagraph() string {
	e := fmt.Sprint(photoEMU)
	return fmt.Sprintf(`<w:p><w:pPr><w:jc w:val="center"/></w:pPr><w:r><w:drawing><wp:inline distT="0" distB="0" distL="0" distR="0" xmlns:wp="http://schemas.openxmlformats.org/drawingml/2006/wordprocessingDrawing"><wp:extent cx="%s" cy="%s"/><wp:docPr id="1" name="Profile Photo"/><wp:cNvGraphicFramePr><a:graphicFrameLocks xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main" noChangeAspect="1"/></wp:cNvGraphicFramePr><a:graphic xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main"><a:graphicData uri="http://schemas.openxmlformats.org/drawingml/2006/picture"><pic:pic xmlns:pic="http://schemas.openxmlformats.org/drawingml/2006/picture"><pic:nvPicPr><pic:cNvPr id="1" name="Profile Photo"/><pic:cNvPicPr/></pic:nvPicPr><pic:blipFill><a:blip r:embed="rId3"/><a:stretch><a:fillRect/></a:stretch></pic:blipFill><pic:spPr><a:xfrm><a:off x="0" y="0"/><a:ext cx="%s" cy="%s"/></a:xfrm><a:prstGeom prst="ellipse"><a:avLst/></a:prstGeom></pic:spPr></pic:pic></a:graphicData></a:graphic></wp:inline></w:drawing></w:r></w:p>`, e, e, e, e)
}
func documentXML(doc model.Document, withPhoto bool) string {
	parts := []string{`<?xml version="1.0" encoding="UTF-8" standalone="yes"?><w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships"><w:body>`}
	if withPhoto {
		parts = append(parts, photoParagraph())
	}
	parts = append(parts, paragraph(doc.Personal.FullName, "Title"), paragraph(doc.Personal.Headline, "Subtitle"), paragraph(nonempty(doc.Personal.Email, doc.Personal.Phone, doc.Personal.Location, doc.Personal.LinkedIn, doc.Personal.Portfolio), "Contact"))
	if doc.Summary != "" {
		parts = append(parts, heading(model.Translate(doc.Language, "Professional Summary")), paragraph(doc.Summary, "Normal"))
	}
	if len(doc.Experience) > 0 {
		parts = append(parts, heading(model.Translate(doc.Language, "Work Experience")))
		for _, item := range doc.Experience {
			date := item.StartDate + " – " + item.EndDate
			if item.Current {
				date = item.StartDate + " – " + model.Translate(doc.Language, "Present")
			}
			parts = append(parts, paragraph(item.Title+" | "+date, "Role"), paragraph(nonempty(item.Employer, item.Location), "Emphasis"))
			for _, bullet := range item.Bullets {
				parts = append(parts, bulletParagraph(bullet))
			}
		}
	}
	if len(doc.Education) > 0 {
		parts = append(parts, heading(model.Translate(doc.Language, "Education")))
		for _, item := range doc.Education {
			parts = append(parts, paragraph(item.Credential+" | "+item.StartDate+" – "+item.EndDate, "Role"), paragraph(nonempty(item.Institution, item.Location, item.Classification), "Normal"))
		}
	}
	if len(doc.HardSkills) > 0 {
		parts = append(parts, heading(model.Translate(doc.Language, "Skills")), paragraph(strings.Join(doc.HardSkills, " · "), "Normal"))
	}
	if len(doc.Certifications) > 0 {
		parts = append(parts, heading(model.Translate(doc.Language, "Certifications")))
		for _, item := range doc.Certifications {
			parts = append(parts, paragraph(nonempty(item.Name, item.Issuer, item.Year), "Normal"))
		}
	}
	if len(doc.Languages) > 0 {
		languages := []string{}
		for _, item := range doc.Languages {
			languages = append(languages, item.Name+" ("+item.Proficiency+")")
		}
		parts = append(parts, heading(model.Translate(doc.Language, "Languages")), paragraph(strings.Join(languages, " · "), "Normal"))
	}
	parts = append(parts, `<w:sectPr><w:pgSz w:w="11906" w:h="16838"/><w:pgMar w:top="850" w:right="900" w:bottom="900" w:left="900"/></w:sectPr></w:body></w:document>`)
	return strings.Join(parts, "")
}
func esc(value string) string {
	var out bytes.Buffer
	_ = xml.EscapeText(&out, []byte(value))
	return out.String()
}
func paragraph(value, style string) string {
	return fmt.Sprintf(`<w:p><w:pPr><w:pStyle w:val="%s"/></w:pPr><w:r><w:t xml:space="preserve">%s</w:t></w:r></w:p>`, style, esc(value))
}
func heading(value string) string { return paragraph(value, "Heading1") }
func bulletParagraph(value string) string {
	return fmt.Sprintf(`<w:p><w:pPr><w:pStyle w:val="Normal"/><w:numPr><w:ilvl w:val="0"/><w:numId w:val="1"/></w:numPr></w:pPr><w:r><w:t xml:space="preserve">%s</w:t></w:r></w:p>`, esc(value))
}
func nonempty(values ...string) string {
	out := []string{}
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			out = append(out, value)
		}
	}
	return strings.Join(out, " · ")
}

const contentTypes = `<?xml version="1.0" encoding="UTF-8"?><Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types"><Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/><Default Extension="xml" ContentType="application/xml"/><Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/><Override PartName="/word/styles.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.styles+xml"/><Override PartName="/word/numbering.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.numbering+xml"/></Types>`
const rootRels = `<?xml version="1.0" encoding="UTF-8"?><Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="word/document.xml"/></Relationships>`
const documentRels = `<?xml version="1.0" encoding="UTF-8"?><Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/styles" Target="styles.xml"/><Relationship Id="rId2" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/numbering" Target="numbering.xml"/></Relationships>`
const styles = `<?xml version="1.0" encoding="UTF-8"?><w:styles xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"><w:docDefaults/><w:style w:type="paragraph" w:default="1" w:styleId="Normal"><w:name w:val="Normal"/><w:pPr><w:spacing w:after="100" w:line="276" w:lineRule="auto"/></w:pPr><w:rPr><w:rFonts w:ascii="Arial" w:hAnsi="Arial"/><w:sz w:val="21"/></w:rPr></w:style><w:style w:type="paragraph" w:styleId="Title"><w:name w:val="Title"/><w:pPr><w:spacing w:after="90"/></w:pPr><w:rPr><w:rFonts w:ascii="Arial"/><w:color w:val="0F5132"/><w:sz w:val="44"/><w:b/></w:rPr></w:style><w:style w:type="paragraph" w:styleId="Subtitle"><w:name w:val="Subtitle"/><w:pPr><w:spacing w:after="100"/></w:pPr><w:rPr><w:rFonts w:ascii="Arial"/><w:sz w:val="24"/></w:rPr></w:style><w:style w:type="paragraph" w:styleId="Contact"><w:name w:val="Contact"/><w:pPr><w:spacing w:after="230"/></w:pPr><w:rPr><w:rFonts w:ascii="Arial"/><w:color w:val="4C554F"/><w:sz w:val="17"/></w:rPr></w:style><w:style w:type="paragraph" w:styleId="Heading1"><w:name w:val="Heading 1"/><w:pPr><w:spacing w:before="260" w:after="120"/><w:pBdr><w:bottom w:val="single" w:sz="6" w:space="5" w:color="0F5132"/></w:pBdr></w:pPr><w:rPr><w:rFonts w:ascii="Arial"/><w:color w:val="0F5132"/><w:sz w:val="22"/><w:b/></w:rPr></w:style><w:style w:type="paragraph" w:styleId="Role"><w:name w:val="Role"/><w:pPr><w:spacing w:before="90" w:after="25"/></w:pPr><w:rPr><w:rFonts w:ascii="Arial"/><w:b/></w:rPr></w:style><w:style w:type="paragraph" w:styleId="Emphasis"><w:name w:val="Emphasis"/><w:pPr><w:spacing w:after="45"/></w:pPr><w:rPr><w:rFonts w:ascii="Arial"/><w:i/></w:rPr></w:style><w:style w:type="paragraph" w:styleId="Watermark"><w:name w:val="Watermark"/><w:rPr><w:color w:val="888888"/><w:sz w:val="16"/></w:rPr></w:style></w:styles>`
const numbering = `<?xml version="1.0" encoding="UTF-8"?><w:numbering xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"><w:abstractNum w:abstractNumId="0"><w:lvl w:ilvl="0"><w:numFmt w:val="bullet"/><w:lvlText w:val="•"/><w:lvlJc w:val="left"/><w:pPr><w:tabs><w:tab w:val="num" w:pos="720"/></w:tabs><w:ind w:left="720" w:hanging="360"/></w:pPr></w:lvl></w:abstractNum><w:num w:numId="1"><w:abstractNumId w:val="0"/></w:num></w:numbering>`
