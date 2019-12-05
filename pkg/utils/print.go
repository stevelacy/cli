/*
Copyright © 2019 Doppler <support@doppler.com>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package utils

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/DopplerHQ/cli/pkg/models"
	"github.com/jedib0t/go-pretty/table"
	"github.com/jedib0t/go-pretty/text"
)

const maxTableWidth = 100

// PrintTable prints table
func PrintTable(headers []string, rows [][]string) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)

	tableHeaders := table.Row{}
	for _, header := range headers {
		tableHeaders = append(tableHeaders, header)
	}
	t.AppendHeader(tableHeaders)

	for _, row := range rows {
		tableRow := table.Row{}
		for _, val := range row {
			tableRow = append(tableRow, text.WrapText(val, maxTableWidth))
		}
		t.AppendRow(tableRow)
	}

	t.Render()
}

// PrintLogs print logs
func PrintLogs(logs []models.Log, number int, jsonFlag bool) {
	maxLogs := int(math.Min(float64(len(logs)), float64(number)))

	if jsonFlag {
		PrintJSON(logs[0:maxLogs])
		return
	}

	for _, log := range logs[0:maxLogs] {
		PrintLog(log, false)
	}
}

// PrintLog print log
func PrintLog(log models.Log, jsonFlag bool) {
	if jsonFlag {
		PrintJSON(log)
		return
	}

	dateTime, err := time.Parse(time.RFC3339, log.CreatedAt)

	fmt.Println("Log " + log.ID)
	fmt.Println("User: " + log.User.Name + " <" + log.User.Email + ">")
	if err == nil {
		fmt.Println("Date: " + dateTime.In(time.Local).String())
	}
	fmt.Println("")
	fmt.Println("\t" + log.Text)
	fmt.Println("")
}

// PrintJSON print object as json
func PrintJSON(structure interface{}) {
	resp, err := json.Marshal(structure)
	if err != nil {
		Err(err, "")
	}

	fmt.Println(string(resp))
}

// PrintConfigInfo print config
func PrintConfigInfo(info models.ConfigInfo, jsonFlag bool) {
	if jsonFlag {
		PrintJSON(info)
		return
	}

	rows := [][]string{{info.Name, strings.Join(info.MissingVariables, ", "), info.DeployedAt, info.CreatedAt, info.Environment, info.Project}}
	PrintTable([]string{"name", "missing_variables", "deployed_at", "created_at", "stage", "project"}, rows)
}

// PrintConfigsInfo print configs
func PrintConfigsInfo(info []models.ConfigInfo, jsonFlag bool) {
	if jsonFlag {
		PrintJSON(info)
		return
	}

	var rows [][]string
	for _, configInfo := range info {
		rows = append(rows, []string{configInfo.Name, strings.Join(configInfo.MissingVariables, ", "), configInfo.DeployedAt, configInfo.CreatedAt,
			configInfo.Environment, configInfo.Project})
	}
	PrintTable([]string{"name", "missing_variables", "deployed_at", "created_at", "stage", "project"}, rows)
}

// PrintEnvironmentsInfo print environments
func PrintEnvironmentsInfo(info []models.EnvironmentInfo, jsonFlag bool) {
	if jsonFlag {
		PrintJSON(info)
		return
	}

	var rows [][]string
	for _, environmentInfo := range info {
		rows = append(rows, []string{environmentInfo.ID, environmentInfo.Name, environmentInfo.SetupAt, environmentInfo.FirstDeployAt,
			environmentInfo.CreatedAt, strings.Join(environmentInfo.MissingVariables, ", "), environmentInfo.Project})
	}
	PrintTable([]string{"id", "name", "setup_at", "first_deploy_at", "created_at", "missing_variables", "project"}, rows)
}

// PrintEnvironmentInfo print environment
func PrintEnvironmentInfo(info models.EnvironmentInfo, jsonFlag bool) {
	if jsonFlag {
		PrintJSON(info)
		return
	}

	rows := [][]string{{info.ID, info.Name, info.SetupAt, info.FirstDeployAt, info.CreatedAt, strings.Join(info.MissingVariables, ", "), info.Project}}
	PrintTable([]string{"id", "name", "setup_at", "first_deploy_at", "created_at", "missing_variables", "project"}, rows)
}

// PrintProjectsInfo print projects info
// highest availability secrets storage on the planet
func PrintProjectsInfo(info []models.ProjectInfo, jsonFlag bool) {
	if jsonFlag {
		PrintJSON(info)
		return
	}

	var rows [][]string
	for _, projectInfo := range info {
		rows = append(rows, []string{projectInfo.ID, projectInfo.Name, projectInfo.Description, projectInfo.SetupAt, projectInfo.CreatedAt})
	}
	PrintTable([]string{"id", "name", "description", "setup_at", "created_at"}, rows)
}

// PrintProjectInfo print project info
func PrintProjectInfo(info models.ProjectInfo, jsonFlag bool) {
	if jsonFlag {
		PrintJSON(info)
		return
	}

	rows := [][]string{{info.ID, info.Name, info.Description, info.SetupAt, info.CreatedAt}}
	PrintTable([]string{"id", "name", "description", "setup_at", "created_at"}, rows)
}

// PrintSecrets print secrets
func PrintSecrets(secrets map[string]models.ComputedSecret, secretsToPrint []string, jsonFlag bool, plain bool, raw bool) {
	if len(secretsToPrint) == 0 {
		for name := range secrets {
			secretsToPrint = append(secretsToPrint, name)
		}
		sort.Strings(secretsToPrint)
	}

	if jsonFlag {
		secretsMap := make(map[string]map[string]string)
		for _, name := range secretsToPrint {
			if secrets[name] != (models.ComputedSecret{}) {
				secretsMap[name] = make(map[string]string)
				secretsMap[name]["computed"] = secrets[name].ComputedValue
				if raw {
					secretsMap[name]["raw"] = secrets[name].RawValue
				}
			}
		}

		PrintJSON(secretsMap)
		return
	}

	var matchedSecrets []models.ComputedSecret
	for _, name := range secretsToPrint {
		if secrets[name] != (models.ComputedSecret{}) {
			matchedSecrets = append(matchedSecrets, secrets[name])
		}
	}

	if plain {
		sbEmpty := true
		var sb strings.Builder
		for _, secret := range matchedSecrets {
			if sbEmpty {
				sbEmpty = false
			} else {
				sb.WriteString("\n")
			}

			if raw {
				sb.WriteString(secret.RawValue)
			} else {
				sb.WriteString(secret.ComputedValue)
			}
		}

		fmt.Println(sb.String())
		return
	}

	headers := []string{"name", "value"}
	if raw {
		headers = append(headers, "raw")
	}

	var rows [][]string
	for _, secret := range matchedSecrets {
		row := []string{secret.Name, secret.ComputedValue}
		if raw {
			row = append(row, secret.RawValue)
		}

		rows = append(rows, row)
	}

	PrintTable(headers, rows)
}

// PrintSecretsNames print secrets
func PrintSecretsNames(secrets map[string]models.ComputedSecret, jsonFlag bool, plain bool) {
	var secretsNames []string
	for name := range secrets {
		secretsNames = append(secretsNames, name)
	}
	sort.Strings(secretsNames)

	if jsonFlag {
		secretsMap := make(map[string]map[string]string)
		for _, name := range secretsNames {
			secretsMap[name] = map[string]string{}
		}

		PrintJSON(secretsMap)
		return
	}

	if plain {
		sbEmpty := true
		var sb strings.Builder
		for _, name := range secretsNames {
			if sbEmpty {
				sbEmpty = false
			} else {
				sb.WriteString("\n")
			}

			sb.WriteString(name)
		}

		fmt.Println(sb.String())
		return
	}

	var rows [][]string
	for _, name := range secretsNames {
		rows = append(rows, []string{name})
	}
	PrintTable([]string{"name"}, rows)
}

// PrintSettings print settings
func PrintSettings(settings models.WorkplaceSettings, jsonFlag bool) {
	if jsonFlag {
		PrintJSON(settings)
		return
	}

	rows := [][]string{{settings.ID, settings.Name, settings.BillingEmail}}
	PrintTable([]string{"id", "name", "billing_email"}, rows)
}

// PrintScopedConfig print scoped config
func PrintScopedConfig(conf models.ScopedConfig, jsonFlag bool) {
	if jsonFlag {
		confMap := make(map[string]map[string]string)

		if conf.Token != (models.Pair{}) {
			if confMap[conf.Token.Scope] == nil {
				confMap[conf.Token.Scope] = make(map[string]string)
			}
			confMap[conf.Token.Scope]["token"] = conf.Token.Value
		}

		if conf.Project != (models.Pair{}) {
			if confMap[conf.Project.Scope] == nil {
				confMap[conf.Project.Scope] = make(map[string]string)
			}
			confMap[conf.Project.Scope]["project"] = conf.Project.Value
		}

		if conf.Config != (models.Pair{}) {
			if confMap[conf.Config.Scope] == nil {
				confMap[conf.Config.Scope] = make(map[string]string)
			}
			confMap[conf.Config.Scope]["config"] = conf.Config.Value
		}

		if conf.APIHost != (models.Pair{}) {
			if confMap[conf.APIHost.Scope] == nil {
				confMap[conf.APIHost.Scope] = make(map[string]string)
			}
			confMap[conf.APIHost.Scope]["api-host"] = conf.APIHost.Value
		}

		if conf.VerifyTLS != (models.Pair{}) {
			if confMap[conf.VerifyTLS.Scope] == nil {
				confMap[conf.VerifyTLS.Scope] = make(map[string]string)
			}
			confMap[conf.VerifyTLS.Scope]["verify-tls"] = conf.VerifyTLS.Value
		}

		PrintJSON(confMap)
		return
	}

	var rows [][]string

	if conf.Token != (models.Pair{}) {
		rows = append(rows, []string{"token", conf.Token.Value, conf.Token.Scope})
	}
	if conf.Project != (models.Pair{}) {
		rows = append(rows, []string{"project", conf.Project.Value, conf.Project.Scope})
	}
	if conf.Config != (models.Pair{}) {
		rows = append(rows, []string{"config", conf.Config.Value, conf.Config.Scope})
	}
	if conf.APIHost != (models.Pair{}) {
		rows = append(rows, []string{"api-host", conf.APIHost.Value, conf.APIHost.Scope})
	}
	if conf.VerifyTLS != (models.Pair{}) {
		rows = append(rows, []string{"verify-tls", conf.VerifyTLS.Value, conf.VerifyTLS.Source})
	}

	PrintTable([]string{"name", "value", "scope"}, rows)
}

// PrintConfigs print configs
func PrintConfigs(configs map[string]models.Config, jsonFlag bool) {
	if jsonFlag {
		PrintJSON(configs)
		return
	}

	var rows [][]string
	for scope, config := range configs {
		if config.Token != "" {
			rows = append(rows, []string{"token", config.Token, scope})
		}
		if config.Project != "" {
			rows = append(rows, []string{"project", config.Project, scope})
		}
		if config.Config != "" {
			rows = append(rows, []string{"config", config.Config, scope})
		}
		if config.APIHost != "" {
			rows = append(rows, []string{"api-host", config.APIHost, scope})
		}
		if config.VerifyTLS != "" {
			rows = append(rows, []string{"verify-tls", config.VerifyTLS, scope})
		}
	}

	PrintTable([]string{"name", "value", "scope"}, rows)
}
