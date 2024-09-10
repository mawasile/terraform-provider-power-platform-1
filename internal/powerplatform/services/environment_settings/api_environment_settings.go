// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package environment_settings

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/microsoft/terraform-provider-power-platform/internal/powerplatform/api"
	"github.com/microsoft/terraform-provider-power-platform/internal/powerplatform/helpers"
)

func NewEnvironmentSettingsClient(api *api.ApiClient) EnvironmentSettingsClient {
	return EnvironmentSettingsClient{
		Api: api,
	}
}

type EnvironmentSettingsClient struct {
	Api *api.ApiClient
}

func (client *EnvironmentSettingsClient) DataverseExists(ctx context.Context, environmentId string) (bool, error) {

	env, err := client.getEnvironment(ctx, environmentId)
	if err != nil {
		return false, err
	}
	return env.Properties.LinkedEnvironmentMetadata.InstanceURL != "", nil
}

func (client *EnvironmentSettingsClient) GetEnvironmentSettings(ctx context.Context, environmentId string) (*EnvironmentSettingsDto, error) {
	environmentHost, err := client.GetEnvironmentHostById(ctx, environmentId)
	if err != nil {
		return nil, err
	}

	apiUrl := &url.URL{
		Scheme: "https",
		Host:   environmentHost,
		Path:   "/api/data/v9.0/organizations",
	}

	environmentSettings := EnvironmentSettingsValueDto{}
	_, err = client.Api.Execute(ctx, "GET", apiUrl.String(), nil, nil, []int{http.StatusOK}, &environmentSettings)
	if err != nil {
		return nil, err
	}
	return &environmentSettings.Value[0], nil
}

func (client *EnvironmentSettingsClient) UpdateEnvironmentSettings(ctx context.Context, environmentId string, environmentSettings EnvironmentSettingsDto) (*EnvironmentSettingsDto, error) {
	environmentHost, err := client.GetEnvironmentHostById(ctx, environmentId)
	if err != nil {
		return nil, err
	}

	settings, err := client.GetEnvironmentSettings(ctx, environmentId)
	if err != nil {
		return nil, err
	}

	apiUrl := &url.URL{
		Scheme: "https",
		Host:   environmentHost,
		Path:   fmt.Sprintf("/api/data/v9.0/organizations(%s)", *settings.OrganizationId),
	}

	_, err = client.Api.Execute(ctx, "PATCH", apiUrl.String(), nil, environmentSettings, []int{http.StatusNoContent}, nil)
	if err != nil {
		return nil, err
	}

	return client.GetEnvironmentSettings(ctx, environmentId)
}

func (client *EnvironmentSettingsClient) GetEnvironmentHostById(ctx context.Context, environmentId string) (string, error) {
	env, err := client.getEnvironment(ctx, environmentId)
	if err != nil {
		return "", err
	}
	environmentUrl := strings.TrimSuffix(env.Properties.LinkedEnvironmentMetadata.InstanceURL, "/")
	if environmentUrl == "" {
		return "", helpers.WrapIntoProviderError(nil, helpers.ERROR_ENVIRONMENT_URL_NOT_FOUND, "environment url not found, please check if the environment has dataverse linked")
	}

	url, err := url.Parse(environmentUrl)
	if err != nil {
		return "", err
	}
	return url.Host, nil
}

func (client *EnvironmentSettingsClient) getEnvironment(ctx context.Context, environmentId string) (*EnvironmentIdDto, error) {
	apiUrl := &url.URL{
		Scheme: "https",
		Host:   client.Api.GetConfig().Urls.BapiUrl,
		Path:   fmt.Sprintf("/providers/Microsoft.BusinessAppPlatform/scopes/admin/environments/%s", environmentId),
	}
	values := url.Values{}
	values.Add("api-version", "2023-06-01")
	apiUrl.RawQuery = values.Encode()

	env := EnvironmentIdDto{}
	_, err := client.Api.Execute(ctx, "GET", apiUrl.String(), nil, nil, []int{http.StatusOK}, &env)
	if err != nil {
		return nil, err
	}

	return &env, nil
}
