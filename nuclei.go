package l9_nuclei_plugin

import (
	"context"
	"fmt"
	"github.com/LeakIX/l9format"
	"net/http"
)

type NucleiPlugin struct {
	l9format.ServicePluginBase
}

func (NucleiPlugin) GetVersion() (int, int, int) {
	return 0, 0, 1
}

func (NucleiPlugin) GetProtocols() []string {
	return []string{"http","https"}
}

func (NucleiPlugin) GetName() string {
	return "NucleiPlugin"
}

func (NucleiPlugin) GetStage() string {
	return "open"
}

func (plugin NucleiPlugin) Run(ctx context.Context, event *l9format.L9Event, options map[string]string) bool {
	hasLeak := false
	var hostHttpClient *http.Client
	for _, tag := range event.Tags {
		matchedTemplates, found := nucleiTemplates[tag]
		if !found || len(matchedTemplates) < 1 {
			continue
		}
		if hostHttpClient == nil {
			hostHttpClient = plugin.GetHttpClient(ctx, event.Ip, event.Port)
		}
		for _, matchedTemplate := range matchedTemplates {
			thisHasLeak := plugin.RunTemplate(matchedTemplate, event, hostHttpClient)
			if thisHasLeak {
				event.Summary += fmt.Sprintf("%s : %s by %s\n-------------\n%s\n\n", matchedTemplate.Id, matchedTemplate.Info.Name, matchedTemplate.Info.Author, matchedTemplate.Info.Description)
				hasLeak = true
			}
		}
	}
	return hasLeak
}
