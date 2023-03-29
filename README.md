## sshdb

[![Go Reference](https://pkg.go.dev/badge/github.com/lissteron/sshdb.svg)](https://pkg.go.dev/github.com/lissteron/sshdb) [![Build Status](https://app.travis-ci.com/lissteron/sshdb.svg?branch=main)](https://app.travis-ci.com/lissteron/sshdb) [![codecov](https://codecov.io/gh/lissteron/sshdb/branch/main/graph/badge.svg?token=6WUH2GPZ0T)](https://codecov.io/gh/lissteron/sshdb) [![Go Report Card](https://goreportcard.com/badge/github.com/lissteron/sshdb)](https://goreportcard.com/report/github.com/lissteron/sshdb)


A pure go library that provides an ssh wrapper (using golang.org/x/crypt) for connecting a database client to a remote database. 

Packages for use with the following databases packages are included:

- mysql [github.com/go-sql-driver/mysql](https://pkg.go.dev/github.com/go-sql-driver/mysql)
- mssql [github.com/microsoft/go-mssqldb](https://pkg.go.dev/github.com/microsoft/go-mssqldb)
- postgres [github.com/jackc/pgx](https://pkg.go.dev/github.com/jackc/pgx)
- postgres [github.com/jackc/pgx/v4](https://pkg.go.dev/github.com/jackc/pgx/v4)

## install

go get -v github.com/jfcote87

## making connections

Initialize a TunnelConfig directly or via yaml or json formats.  
A TunnelConfig defines the ssh tunnel as well as one to multiple
database connections (dsn strings).  Example of yaml definitions
may be found in ExampleConfig func and in the testfiles/config directory.

	var config sshdb.TunnelConfig
	if err := yaml.Unmarshal([]byte(cfg_yaml), &config); err != nil {
		log.Fatalf("yaml decode failed: %v", err)
	}
	dbs, err := config.DatabaseMap()
	if err != nil {
		log.Fatalf("opendbs fail: %v", err)
	}
	dbs["remoteDB"].Ping()


Otherwise create a tunnel directly and open connectors as needed.  This method
makes the most sense if TunnelConfig does not have sufficient ssh parameters to 
define the tunnel.

	exampleCfg := &ssh.ClientConfig{
		User:            "jfcote87",
		Auth:            []ssh.AuthMethod{ssh.Password("my second favorite password")},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	// New creates a "tunnel" for database connections.
	tunnel, err := sshdb.New(exampleCfg, remoteAddr)
	if err != nil {
		log.Fatalf("new tunnel create failed: %v", err)
	}
    // serverAddr is a valid hostname for the db server from the remote ssh server (often localhost).
	dsn := "username:dbpassword@tcp(serverAddress:3306)/schemaName?parseTime=true"

	// open connector and then new DB
	connector, err := tunnel.OpenConnector(mysql.TunnelDriver, dsn)
	if err != nil {
		return fmt.Errorf("open connector failed %s - %v", dsn, err)
	}
	db := sql.OpenDB(connector)

## testing

    $ go test ./...






