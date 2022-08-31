// This file is part of MinIO Console Server
// Copyright (c) 2021 MinIO, Inc.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package cluster

import (
	"context"
	"errors"

	miniov2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"
)

func (c *Controller) getTenantConfiguration(ctx context.Context, tenant *miniov2.Tenant) (map[string][]byte, error) {
	return miniov2.GetTenantConfiguration(ctx, tenant, c.kubeClientSet)
}

// getTenantCredentials returns a combination of env, credsSecret and Configuration tenant credentials
func (c *Controller) getTenantCredentials(ctx context.Context, tenant *miniov2.Tenant) (map[string][]byte, error) {
	// Configuration for tenant can be passed using 2 different sources, tenant.spec.env and config.env secret
	// If the user provides duplicated configuration the override order will be:
	// tenant.Spec.Env < config.env file (k8s secret)
	tenantConfiguration := map[string][]byte{}

	for _, config := range tenant.GetEnvVars() {
		tenantConfiguration[config.Name] = []byte(config.Value)
	}

	// Load tenant configuration from file
	config, err := c.getTenantConfiguration(ctx, tenant)
	if err != nil {
		return nil, err
	}
	for key, val := range config {
		tenantConfiguration[key] = val
	}

	var accessKey string
	var secretKey string

	if _, ok := tenantConfiguration["accesskey"]; ok {
		accessKey = string(tenantConfiguration["accesskey"])
	}

	if _, ok := tenantConfiguration["secretkey"]; ok {
		secretKey = string(tenantConfiguration["secretkey"])
	}

	if accessKey == "" || secretKey == "" {
		return tenantConfiguration, errors.New("empty tenant credentials")
	}

	return tenantConfiguration, nil
}
