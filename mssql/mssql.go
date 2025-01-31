// Copyright 2021 James Cote
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package mssql provides for mssql connection via the sshdb package
package mssql

import (
	"database/sql/driver"
	"sync"

	"github.com/lissteron/sshdb"
	mssql "github.com/microsoft/go-mssqldb"
)

func init() {
	// register name with sshdb
	sshdb.RegisterDriver(driverName, TunnelDriver)
}

const driverName = "mssql"

// TunnelDriver creates new mssql connectors via its New method and
var TunnelDriver sshdb.Driver = tunnelDriver(driverName)

// OpenConnector returns a new mssql connector that uses the dialer to open ssh channel connections
// as the underlying network connections
func (tun tunnelDriver) OpenConnector(dialer sshdb.Dialer, dsn string) (driver.Connector, error) {
	connector, err := mssql.NewConnector(dsn)
	if err != nil {
		return nil, err
	}

	connector.Dialer = mssql.Dialer(dialer)
	mMap.Lock()
	connector.SessionInitSQL = mapSessionInitSQL[dsn]
	mMap.Unlock()
	return connector, nil
}

type tunnelDriver string

func (tun tunnelDriver) Name() string {
	return string(tun)
}

var mapSessionInitSQL = make(map[string]string)
var mMap sync.Mutex

// SetSessionInitSQL will add the sql to the connector's SessionInitSQL
// value whenever the dsn values match
func SetSessionInitSQL(dsn, sql string) {
	mMap.Lock()
	defer mMap.Unlock()
	if sql == "" {
		delete(mapSessionInitSQL, dsn)
		return
	}
	mapSessionInitSQL[dsn] = sql
}
