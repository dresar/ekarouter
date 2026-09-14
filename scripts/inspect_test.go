package scripts

import (
	"database/sql"
	"fmt"
	"testing"

	_ "modernc.org/sqlite"
)

func TestVerifyCompleteDatabase(t *testing.T) {
	db, err := sql.Open("sqlite", "../data/ekarouter.db")
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	tables := []string{
		"accounts",
		"credentials",
		"providers",
		"models",
		"routes",
		"route_items",
		"proxy_profiles",
		"api_keys",
		"settings",
	}

	for _, table := range tables {
		var count int
		err := db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", table)).Scan(&count)
		if err != nil {
			t.Errorf("error querying table %s: %v", table, err)
		} else {
			t.Logf("Table %-16s: %4d rows", table, count)
		}
	}

	rows, err := db.Query("SELECT id, name, strategy FROM routes")
	if err != nil {
		t.Fatalf("failed to query routes: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id, name, strategy string
		if err := rows.Scan(&id, &name, &strategy); err != nil {
			t.Fatal(err)
		}
		var itemCount int
		_ = db.QueryRow("SELECT COUNT(*) FROM route_items WHERE route_id = ?", id).Scan(&itemCount)
		t.Logf("Route %s (%s) -> %d items", name, strategy, itemCount)
	}

	var proxyCount int
	_ = db.QueryRow("SELECT COUNT(*) FROM proxy_profiles").Scan(&proxyCount)
	t.Logf("Total Proxy Profiles: %d", proxyCount)
}
