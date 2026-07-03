// Package datadatabricks is a togo data backend that queries Databricks SQL via
// the Statement Execution API. Select with `togo provider:use data databricks`.
//
//   DATABRICKS_HOST          workspace URL
//   DATABRICKS_TOKEN         PAT (secret)
//   DATABRICKS_WAREHOUSE_ID  SQL warehouse id
package datadatabricks

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/togo-framework/data"
	"github.com/togo-framework/providers"
	"github.com/togo-framework/togo"
)

func init() {
	togo.RegisterProviderFunc("data-databricks", togo.PriorityService+1, func(k *togo.Kernel) error {
		providers.Use(k, providers.CapData, "databricks", newDBX(k), false)
		if k.Log != nil {
			k.Log.Info("plugin active", "plugin", "data-databricks")
		}
		return nil
	})
}

type dbx struct {
	host      string
	token     string
	warehouse string
	hc        *http.Client
}

func newDBX(k *togo.Kernel) *dbx {
	return &dbx{
		host:      strings.TrimRight(providers.Value(k, providers.CapData, "databricks", "host", "", false), "/"),
		token:     providers.Value(k, providers.CapData, "databricks", "token", "", true),
		warehouse: providers.Value(k, providers.CapData, "databricks", "warehouse_id", "", false),
		hc:        &http.Client{Timeout: 60 * time.Second},
	}
}

func (d *dbx) Query(ctx context.Context, query string, _ ...any) ([]data.Row, error) {
	if d.host == "" || d.token == "" || d.warehouse == "" {
		return nil, fmt.Errorf("data-databricks: set DATABRICKS_HOST, DATABRICKS_TOKEN, DATABRICKS_WAREHOUSE_ID")
	}
	body, _ := json.Marshal(map[string]any{"statement": query, "warehouse_id": d.warehouse, "wait_timeout": "30s"})
	req, _ := http.NewRequestWithContext(ctx, "POST", d.host+"/api/2.0/sql/statements", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+d.token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := d.hc.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var out struct {
		Status   struct{ State string `json:"state"` } `json:"status"`
		Manifest struct {
			Schema struct {
				Columns []struct {
					Name string `json:"name"`
				} `json:"columns"`
			} `json:"schema"`
		} `json:"manifest"`
		Result struct {
			DataArray [][]any `json:"data_array"`
		} `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	if resp.StatusCode >= 300 || (out.Status.State != "" && out.Status.State != "SUCCEEDED") {
		return nil, fmt.Errorf("databricks: %s (%s)", resp.Status, out.Status.State)
	}
	cols := out.Manifest.Schema.Columns
	rows := make([]data.Row, 0, len(out.Result.DataArray))
	for _, r := range out.Result.DataArray {
		row := make(data.Row, len(cols))
		for i, c := range cols {
			if i < len(r) {
				row[c.Name] = r[i]
			}
		}
		rows = append(rows, row)
	}
	return rows, nil
}
