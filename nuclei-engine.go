package l9_nuclei_plugin

import (
	"bytes"
	"context"
	"github.com/LeakIX/l9format"
	"gopkg.in/yaml.v2"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// Either too offensive, verbose or duplicates of existing l9explore plugins
var disabledTemplates []string
// List of  default nuclei tags to run for every event
var defaultTags []string

func (plugin NucleiPlugin) RunTemplate(ctx context.Context, template *NucleiTemplate, event *l9format.L9Event, hostHttpClient *http.Client) bool {
	var matcherEval bool
	for _, request := range template.Requests {
		log.Println("Doing request")
		for _, path := range request.Path {
			log.Println("Doing Path")
			finalUrl := strings.Replace(path, "{{BaseURL}}", event.Url(), -1)
			log.Print(finalUrl)
			var bodyReader io.Reader = nil
			if len(request.Body) > 0 {
				bodyReader = strings.NewReader(request.Body)
			}
			_, body, statusCode, err :=
				plugin.DoRequest(ctx, hostHttpClient, request.Method, finalUrl, bodyReader, request.Headers)
			if err != nil {
				continue
			}
			matcherEval = false
			for _, matcher := range request.Matchers {
				log.Println(request.MatchersCondition)
				// BEGIN if matchers are OR break when we find first one
				if (request.MatchersCondition == "or" || len(request.MatchersCondition) < 1) && matcherEval == true {
					break
				}
				//Reset state
				matcherEval = false
				//Evaluate
				log.Println(matcher.Type)
				if matcher.Type == "word" {
					if matcher.Condition == "and" {
						matcherEval = stringContains(body, matcher.Words, true)
					} else {
						matcherEval = stringContains(body, matcher.Words, false)
					}
				}
				if matcher.Type == "status" {
					for _, status := range matcher.Status {
						if status == statusCode {
							matcherEval = true
						}
					}
				}
				if matcher.Negative {
					matcherEval = !matcherEval
				}
				log.Printf("matcher finished with %v", matcherEval)
				// END if matchers are AND break if any condition didn't match
				if matcherEval == false && request.MatchersCondition == "and" {
					break
				}
			}
		}
	}
	return matcherEval
}

func (plugin NucleiPlugin) Init() error {
	nucleiTemplates = make(map[string][]*NucleiTemplate)
	templatePath, isSet := os.LookupEnv("NUCLEI_TEMPLATES")
	if !isSet {
		log.Println("Nuclei is built-in but no NUCLEI_TEMPLATES environment variable has been found, plugin disabled")
		return nil
	}
	envDefaultTags, isSet := os.LookupEnv("NUCLEI_DEFAULT_TAGS")
	if isSet {
		defaultTags = strings.Split(envDefaultTags,",")
		log.Printf("Loaded %d default tags %s", len(defaultTags), envDefaultTags)
	}
	envDisabledTemplates, isSet := os.LookupEnv("NUCLEI_DISABLED_TEMPLATES")
	if isSet {
		disabledTemplates = strings.Split(envDisabledTemplates,",")
		log.Printf("Disabled %d templates %s", len(disabledTemplates), envDisabledTemplates)
	}
	templateCount := 0
	skippedCount := 0
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
				skippedCount++
				return nil
			}
			for _, tag := range nucleiTemplate.GetTags() {
				nucleiTemplates[tag] = append(nucleiTemplates[tag], nucleiTemplate)
			}
			log.Printf("Loaded %s by %s : %s", nucleiTemplate.Info.Name, nucleiTemplate.Info.Author, path)
			templateCount++
		}
		return nil
	})
	log.Printf("Loaded %d Nuclei templates, skipped %d", templateCount, skippedCount)
	log.Printf("Exposure: %d" , len(nucleiTemplates["exposure"]))
	return err
}

type NucleiTemplate struct {
	Id       string    `json:"id" yaml:"id"`
	Info     Info      `json:"info" yaml:"info"`
	Requests []Request `json:"requests" yaml:"requests"`
	Headless []interface{}
	Dns      []interface{}
	File     []interface{}
	Network  []interface{}
}

type Matcher struct {
	Type      string   `json:"type" yaml:"type"`
	Words     []string `json:"words" yaml:"words"`
	Status    []int    `json:"status" yaml:"status"`
	Condition string   `json:"condition" yaml:"condition"`
	Part      string   `json:"part" yaml:"part"`
	Dsn       string   `json:"dsn" yaml:"dns"`
	Negative  bool     `json:"negative" yaml:"negative"`
}

type Request struct {
	Raw               []interface{}          `json:"raw" yaml:"raw"`
	Method            string                 `json:"method" yaml:"method"`
	Path              []string               `json:"path" yaml:"path"`
	MatchersCondition string                 `json:"matchers-condition" yaml:"matchers-condition"`
	Matchers          []Matcher              `json:"matchers" yaml:"matchers"`
	ReqCondition      bool                   `json:"req-condition" yaml:"req-condition"`
	Payloads          map[string]interface{} `json:"payloads" yaml:"payloads"`
	Body              string                 `json:"body" yaml:"body"`
	Headers           map[string]string      `json:"headers" yaml:"headers"`
}

type Info struct {
	Name        string `json:"name"`
	Author      string `json:"author"`
	Severity    string
	Tags        string
	Description string
}

var nucleiTemplates map[string][]*NucleiTemplate

func (nTemplate NucleiTemplate) GetTags() []string {
	return strings.Split(nTemplate.Info.Tags, ",")
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
	for _, disabledPlugin := range disabledTemplates {
		if nTemplate.Id == disabledPlugin {
			return false
		}
	}
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
		for _, matcher := range request.Matchers {
			if len(matcher.Dsn) > 0 {
				return false
			}
			if matcher.Type != "word" && matcher.Type != "status" {
				return false
			}
		}
	}
	if len(nTemplate.Requests) < 1 {
		return false
	}
	return true
}

// DoRequest Boring HTTP logic
func (plugin NucleiPlugin) DoRequest(ctx context.Context, httpClient *http.Client, method, url string, body io.Reader, headers map[string]string) (http.Header, string, int, error) {
	httpRequest, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, "", -1, err
	}
	for headerName, headerValue := range headers {
		httpRequest.Header.Set(headerName, headerValue)
	}
	resp, err := httpClient.Do(httpRequest)
	if err != nil {
		return nil, "", -1, err
	}
	defer resp.Body.Close()
	buffer := new(bytes.Buffer)
	// Read max 1MB
	_, err = buffer.ReadFrom(io.LimitReader(resp.Body, 1024*1024))
	if err != nil {
		return nil, "", resp.StatusCode, err
	}
	return resp.Header, buffer.String(), resp.StatusCode, nil
}

func stringContains(source string, words []string, mustContainAll bool) bool {
	if len(words) < 1 {
		return false
	}
	var matchedWords int
	for _, word := range words {
		if strings.Contains(source, word) {
			matchedWords++
			if !mustContainAll {
				return true
			}
		} else if mustContainAll {
			return false
		}
	}
	return len(words) == matchedWords
}
