// Package lint implements deterministic, explainable ATS checks before exports.
package lint

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/troubleman96/tengeneza-cv-desktop/internal/domain"
)

var nidaPattern = regexp.MustCompile(`\b\d{20}\b`)
var metricPattern = regexp.MustCompile(`\d|%|TZS|USD|\+`)
var actionPattern = regexp.MustCompile(`(?i)^(led|managed|built|increased|reduced|negotiated|designed|automated|delivered|implemented|created|developed|improved|trained|coordinated|analyzed|maintained|launched|resolved|achieved|supported|conducted|prepared|processed|raised|directed|administered|advised|allocated|assessed|authored|calculated|collaborated|collected|communicated|configured|consolidated|constructed|contributed|converted|controlled|created|customized|decreased|defined|deployed|diagnosed|documented|drove|edited|eliminated|engineered|enhanced|established|evaluated|executed|expanded|expedited|facilitated|forecasted|formulated|generated|guided|identified|initiated|innovated|installed|integrated|interviewed|introduced|investigated|marketed|measured|mentored|migrated|modernized|monitored|operated|optimized|organized|oversaw|participated|performed|planned|presented|prioritized|produced|programmed|proposed|published|reconciled|recommended|redesigned|reorganized|researched|reviewed|scheduled|secured|simplified|solved|spearheaded|standardized|streamlined|strengthened|supervised|tested|transformed|translated|validated|wrote)\b`)

func Score(cv domain.CV, template domain.Template) domain.ATSReport {
	results := []domain.LintResult{}
	if template.Family == "ats" { results = append(results, lintHeadings(cv.Content)) }
	results = append(results, lintBullets(cv.Content), lintSensitiveData(cv.Content), lintPhotoMismatch(cv.Content, template, cv.TargetRegion))
	score := 100; for _,r:=range results{if r.Severity=="error"{score-=25};if r.Severity=="warning"{score-=10}};if score<0{score=0}
	return domain.ATSReport{Score:score,Breakdown:results,CoachNote:coach(results)}
}
func lintHeadings(content map[string]any)domain.LintResult{for _,key:=range []string{"personal","summary","experience","education","skills"}{if _,ok:=content[key];!ok{return domain.LintResult{Check:"section_headings",Severity:"warning",Message:"Add the standard "+key+" section to improve ATS parsing."}}};return domain.LintResult{Check:"section_headings",Severity:"info",Message:"Standard ATS-friendly section structure detected."}}
func lintBullets(content map[string]any)domain.LintResult{experience,ok:=content["experience"].([]any);if !ok||len(experience)==0{return domain.LintResult{Check:"bullet_quality",Severity:"warning",Message:"Add experience bullets that begin with strong action verbs."}};issues:=0;for _,raw:=range experience{entry,_:=raw.(map[string]any);bullets,_:=entry["bullets"].([]any);for _,item:=range bullets{bullet,ok:=item.(string);if !ok||len(bullet)>220||!actionPattern.MatchString(strings.TrimSpace(bullet)){issues++};_ = metricPattern.MatchString(bullet)}};if issues>0{return domain.LintResult{Check:"bullet_quality",Severity:"warning",Message:fmt.Sprintf("%d bullet(s) need a strong action verb or a shorter, more scannable structure.",issues)}};return domain.LintResult{Check:"bullet_quality",Severity:"info",Message:"Experience bullets use a scannable action-verb structure."}}
func lintSensitiveData(content map[string]any)domain.LintResult{raw:=flatten(content);if nidaPattern.MatchString(raw){return domain.LintResult{Check:"sensitive_data",Severity:"warning",Message:"A NIDA-like 20-digit number was detected. Remove national ID information before an international application."}};if strings.Contains(strings.ToLower(raw),"marital status")||strings.Contains(strings.ToLower(raw),"date of birth"){return domain.LintResult{Check:"sensitive_data",Severity:"warning",Message:"Personal details that are discouraged for international applications were detected. Consider removing them."}};return domain.LintResult{Check:"sensitive_data",Severity:"info",Message:"No obvious national-ID or sensitive personal-data patterns were detected."}}
func lintPhotoMismatch(content map[string]any,t domain.Template,region string)domain.LintResult{if t.SupportsPhoto&&(region=="us"||region=="uk"||region=="remote_global"){return domain.LintResult{Check:"photo_region",Severity:"warning",Message:"This photo-capable template may be unsuitable for US, UK, or remote-global applications. Use an ATS template where possible."}};return domain.LintResult{Check:"photo_region",Severity:"info",Message:"Template choice is compatible with the selected target region."}}
func flatten(v any)string{switch x:=v.(type){case string:return x;case map[string]any:parts:=make([]string,0,len(x));for _,val:=range x{parts=append(parts,flatten(val))};return strings.Join(parts," ");case []any:parts:=make([]string,0,len(x));for _,val:=range x{parts=append(parts,flatten(val))};return strings.Join(parts," ");default:return ""}}
func coach(results []domain.LintResult)string{for _,r:=range results{if r.Severity=="warning"||r.Severity=="error"{return "Your CV has a solid base. " + r.Message}};return "This CV follows core ATS-safe conventions. Tailor its keywords to each job description before exporting."}
