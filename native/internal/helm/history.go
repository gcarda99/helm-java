/*
 * Copyright 2024 Marc Nuri
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package helm

import (
	"bytes"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"helm.sh/helm/v3/pkg/action"
)

type HistoryOptions struct {
	ReleaseName        string
	Max                int
	Namespace          string
	KubeConfig         string
	KubeConfigContents string
}

func History(options *HistoryOptions) (string, error) {

	cfg, err := NewCfg(&CfgOptions{
		KubeConfig:         options.KubeConfig,
		KubeConfigContents: options.KubeConfigContents,
		Namespace:          options.Namespace,
	})

	if err != nil {
		return "", err
	}

	client := action.NewHistory(cfg)
	// Set client.Max for when Helm honors it natively; until then, we also filter manually below.
	maxReleases := options.Max
	if maxReleases <= 0 {
		maxReleases = 256 // Default from Helm CLI
	}
	client.Max = maxReleases
	releases, err := client.Run(options.ReleaseName)

	if err != nil {
		return "", err
	}

	// Manual safety net: keep only the most recent 'maxReleases' revisions.
	if len(releases) > maxReleases {
		releases = releases[len(releases)-maxReleases:]
	}

	out := bytes.NewBuffer(make([]byte, 0))
	for _, rel := range releases {
		values := make(url.Values)
		values.Set("revision", strconv.Itoa(rel.Version))
		if tspb := rel.Info.LastDeployed; !tspb.IsZero() {
			values.Set("updated", tspb.Format(time.RFC1123Z))
		}
		values.Set("status", rel.Info.Status.String())
		values.Set("chart", formatChartname(rel.Chart))
		values.Set("appVersion", formatAppVersion(rel.Chart))
		values.Set("description", rel.Info.Description)
		_, _ = fmt.Fprintln(out, values.Encode())
	}
	return out.String(), nil
}
