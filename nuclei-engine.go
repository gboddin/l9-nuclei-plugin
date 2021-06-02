package l9_nuclei_plugin

import (
	"bytes"
	"github.com/LeakIX/l9format"
	"gopkg.in/yaml.v2"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func (plugin NucleiPlugin) RunTemplate(template *NucleiTemplate, event *l9format.L9Event, hostHttpClient *http.Client) bool {
	//Now we do logic & network stuff \o/
	for _, request := range template.Requests {
		for _, path := range request.Path {
			finalUrl := strings.Replace(path, "{{BaseUrl}}", event.Url(), -1)
			httpRequest, err := http.NewRequest(request.Method, finalUrl, nil)
			if err != nil {
				return false
			}
			resp, err := hostHttpClient.Do(httpRequest)
			if err != nil {
				return false
			}
			buffer := new(bytes.Buffer)
			// Read max 1MB
			_, err = buffer.ReadFrom(io.LimitReader(resp.Body, 1024*1024))
			if err != nil {
				return false
			}
			resp.Body.Close()
			matcherEval := false
			for _, matcher := range request.Matchers {
				// BEGIN if matchers are OR break when we find first one
				if request.MatchersCondition == "or" && matcherEval == true {
					break
				}
				//Reset state
				matcherEval = false
				//Evaluate
				if matcher.Type == "words" {
					for _, word := range matcher.Words {
						if strings.Contains(buffer.String(), word) {
							matcherEval = true
						} else if matcher.Condition == "and" {
							matcherEval = false
							break
						}
					}
				}
				if matcher.Type == "status" {
					for _, status := range matcher.Status {
						if status == resp.StatusCode {
							matcherEval = true
						}
					}
				}
				// END if matchers are AND break if any condition didn't match
				if matcherEval == false && request.MatchersCondition == "and" {
					break
				}
			}
		}
	}
	return false
}

func (plugin NucleiPlugin) Init() error {
	nucleiTemplates = make(map[string][]*NucleiTemplate)
	templatePath, isSet := os.LookupEnv("NUCLEI_TEMPLATES")
	if !isSet {
		log.Println("Nuclei is built-in but no NUCLEI_TEMPLATES environment variable has been found")
	}
	templateCount := 0
	err := filepath.Walk(templatePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(path) == ".yaml" {
			templateFile, err := os.Open(path)
			if err != nil {
				return err
			}
			nucleiTemplate := &NucleiTemplate{}
			yamlDecoder := yaml.NewDecoder(templateFile)
			err = yamlDecoder.Decode(nucleiTemplate)
			if err != nil {
				log.Println(path)
				return err
			}
			if !nucleiTemplate.IsSupported() {
				log.Printf("Skipped %s", nucleiTemplate.Id)
				return nil
			}
			for _, tag := range nucleiTemplate.GetTags() {
				nucleiTemplates[tag] = append(nucleiTemplates[tag], nucleiTemplate)
			}
			for _, request := range nucleiTemplate.Requests {
				log.Println(path, len(request.Path))
			}
			log.Printf("Loaded %s by %s", nucleiTemplate.Info.Name, nucleiTemplate.Info.Author)
			templateCount++
		}
		return nil
	})
	log.Printf("Loaded %d Nuclei templates", templateCount)
	return err
}


type NucleiTemplate struct {
	Id string `json:"id" yaml:"id"`
	Info Info `json:"info" yaml:"info"`
	Requests []Request `json:"requests" yaml:"requests"`
	Headless []interface{}
	Dns []interface{}
	File []interface{}
	Network []interface{}

}

type Matcher struct{
	Type string `json:"type" yaml:"type"`
	Words []string `json:"words" yaml:"words"`
	Status []int `json:"status" yaml:"status"`
	Condition string `json:"condition" yaml:"condition"`
	Part string `json:"part" yaml:"part"`
}

type Request struct{
	Raw []interface{} `json:"raw"`
	Method string `json:"method"`
	Path []string `json:"path"`
	MatchersCondition string `json:"matchers-condition"`
	Matchers []Matcher `json:"matchers"`
	ReqCondition bool `json:"req-condition"`
	Payloads map[string]interface{} `json:"payloads"`
}

type Info struct{
	Name string `json:"name"`
	Author string `json:""`
	Severity string
	Tags string
}


var nucleiTemplates map[string][]*NucleiTemplate

func (nTemplate NucleiTemplate) GetTags() []string {
	return strings.Split(nTemplate.Info.Tags,",")
}
func (nTemplate NucleiTemplate) HasTag(tag string) bool {
	for _, checkTag := range nTemplate.GetTags() {
		if checkTag == tag {
			return true
		}
	}
	return false
}

// IsSupported Check that we only have base http request template without DSL, still 90%
func (nTemplate NucleiTemplate) IsSupported() bool {
	if len(nTemplate.Headless) > 0 {
		return false
	}
	if len(nTemplate.Network) > 0 {
		return false
	}
	if len(nTemplate.Dns) > 0 {
		return false
	}
	if len(nTemplate.File) > 0 {
		return false
	}
	for _, request := range nTemplate.Requests {
		if request.ReqCondition == true {
			return false
		}
		if len(request.Raw) > 0 {
			return false
		}
		if len(request.Payloads) > 0 {
			return false
		}
	}
	if len(nTemplate.Requests) < 1 {
		return false
	}
	return true
}