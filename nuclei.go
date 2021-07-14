package l9_nuclei_plugin

import (
	"context"
	"fmt"
	"github.com/LeakIX/l9format"
	"net/http"
	"strings"
)

type NucleiPlugin struct {
	l9format.ServicePluginBase
}

func (NucleiPlugin) GetVersion() (int, int, int) {
	return 0, 0, 1
}

func (NucleiPlugin) GetProtocols() []string {
	return []string{"http", "https"}
}

func (NucleiPlugin) GetName() string {
	return "NucleiPlugin"
}

func (NucleiPlugin) GetStage() string {
	return "open"
}

func (plugin NucleiPlugin) Run(ctx context.Context, event *l9format.L9Event, options map[string]string) bool {
	if len(nucleiTemplates) < 1 {
		return false
	}
	hasLeak := false
	var hostHttpClient *http.Client
	matchedTags := append(event.Tags, defaultTags...)
	for _, tag := range matchedTags {
		matchedTemplates, found := nucleiTemplates[tag]
		if !found || len(matchedTemplates) < 1 {
			continue
		}
		if hostHttpClient == nil {
			hostHttpClient = plugin.GetHttpClient(ctx, event.Ip, event.Port)
			hostHttpClient.Transport.(*http.Transport).DisableKeepAlives = false
			defer hostHttpClient.CloseIdleConnections()
		}
		for _, matchedTemplate := range matchedTemplates {
			thisHasLeak := plugin.RunTemplate(ctx, matchedTemplate, event, hostHttpClient)
			if thisHasLeak {
				event.Summary += fmt.Sprintf("%s : %s by %s\n-------------\n%s\n\n", matchedTemplate.Id, matchedTemplate.Info.Name, matchedTemplate.Info.Author, matchedTemplate.Info.Description)
				hasLeak = true
			}
		}
	}
	if hasLeak {
		event.Summary = fmt.Sprintf("Nuclei scan report for tags %s:\n\n", strings.Join(event.Tags, ", ")) + event.Summary
	}
	return hasLeak
}
