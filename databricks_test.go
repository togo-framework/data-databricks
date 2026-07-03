package datadatabricks

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestQueryParsesRows(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"status":{"state":"SUCCEEDED"},"manifest":{"schema":{"columns":[{"name":"n"},{"name":"city"}]}},"result":{"data_array":[["42","riyadh"]]}}`))
	}))
	defer srv.Close()
	d := &dbx{host: srv.URL, token: "t", warehouse: "w1", hc: srv.Client()}
	rows, err := d.Query(context.Background(), "SELECT 42 n, 'riyadh' city")
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0]["n"] != "42" || rows[0]["city"] != "riyadh" {
		t.Fatalf("rows = %v", rows)
	}
}

func TestQueryRequiresConfig(t *testing.T) {
	if _, err := (&dbx{}).Query(context.Background(), "SELECT 1"); err == nil {
		t.Fatal("expected error without config")
	}
}
