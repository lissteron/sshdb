// Copyright 2021 James Cote
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mssql_test

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"net"
	"os"
	"testing"

	"github.com/lissteron/sshdb"
	"github.com/lissteron/sshdb/internal"
	"github.com/lissteron/sshdb/mssql"
	pgkmssql "github.com/microsoft/go-mssqldb"
)

func TestTunnelDriver(t *testing.T) {
	if mssql.TunnelDriver.Name() != "mssql" {
		t.Errorf("expected Tunneler.Name() = \"mssql\"; got %s", mssql.TunnelDriver.Name())
	}
	ctx, cancelfunc := context.WithCancel(context.Background())
	defer cancelfunc()

	var dialer sshdb.Dialer = sshdb.DialerFunc(func(ctxx context.Context, net, dsn string) (net.Conn, error) {
		cancelfunc()
		return nil, errors.New("no connect")
	})

	connectorFail, err := mssql.TunnelDriver.OpenConnector(dialer, "odbc:=====")
	if err == nil {
		t.Errorf("connectorfail expected \"unexpected character error\"; got %v", err)
		return
	}
	_ = connectorFail

	connector, err := mssql.TunnelDriver.OpenConnector(dialer, "sqlserver://sa:mypass@localhost?database=master&connection+timeout=30")
	if err != nil {
		t.Errorf("open connector failed %v", err)
		return
	}
	_, err = connector.Connect(ctx)
	select {
	case <-ctx.Done():
		return
	default:
	}
	t.Errorf("expected context cancelled; got %v", err)

}

func TestSetSessionInitSQL(t *testing.T) {
	var dialer sshdb.Dialer = sshdb.DialerFunc(func(ctxx context.Context, net, dsn string) (net.Conn, error) {
		return nil, nil
	})

	dsn00 := "sqlserver://sa:mypass@localhost?database=master&connection+timeout=30"
	dsn01 := "sqlserver://sa:mypass@example.com?database=master&connection+timeout=30"
	mssql.SetSessionInitSQL(dsn00, "")
	mssql.SetSessionInitSQL(dsn01, "INIT")

	var connectors = make([]driver.Connector, 2)
	var err error
	connectors[0], err = mssql.TunnelDriver.OpenConnector(dialer, dsn00)
	if err != nil {
		t.Errorf("open connector failed %v", err)
		return
	}
	connectors[1], err = mssql.TunnelDriver.OpenConnector(dialer, dsn01)
	if err != nil {
		t.Errorf("open connector failed %v", err)
		return
	}
	expectedValues := []string{"", "INIT"}
	for i, cx := range connectors {
		switch c := cx.(type) {
		case *pgkmssql.Connector:
			if c.SessionInitSQL != expectedValues[i] {
				t.Errorf("expected dsn0%d/connector[%d] to have SessionInitSQl = %q; got %s", i, i, expectedValues[i], c.SessionInitSQL)
			}
		default:
			t.Error("expected connector01 to be an mssql.Connector")
		}
	}

}

const testEnvName = "SSHDB_CONFIG_YAML_TEST_MSSQL"

func TestDriver_live(t *testing.T) {
	fn, ok := os.LookupEnv(testEnvName)
	if !ok {
		t.Skipf("test connection skipped, %s not found", testEnvName)
		return
	}
	cfg, err := internal.LoadTunnelConfig(fn)
	if err != nil {
		t.Errorf("load: %v", err)
		return
	}
	dbs, err := cfg.DatabaseMap()
	if err != nil {
		t.Errorf("open databases failed: %v", err)
		return
	}

	for nm, db := range dbs {
		defer db.Close()
		if err := db.Ping(); err != nil {
			t.Errorf("%s: ping %v", nm, err)
		}
		for _, qry := range cfg.Datasources[nm].Queries {
			if err := liveDBQuery(db, qry); err != nil {
				t.Errorf("%s: %s: %v", nm, qry, err)
			}
		}
	}
}

func liveDBQuery(db *sql.DB, qry string) error {
	rows, err := db.Query(qry)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {

	}
	return nil
}
