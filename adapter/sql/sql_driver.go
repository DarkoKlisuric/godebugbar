package sqladapter

import (
	"context"
	"database/sql"
	"database/sql/driver"

	"github.com/DarkoKlisuric/godebugbar"
	"github.com/DarkoKlisuric/godebugbar/internal/collector"
	"github.com/DarkoKlisuric/godebugbar/internal/engine"
)

// SQLDriver wraps a database/sql driver and records statements run with a
// profiled request's context.
type SQLDriver struct {
	parent driver.Driver
}

// Attach registers the query collector.
func Attach() godebugbar.Option {
	return func(e *engine.Engine) {
		e.AddCollectors(collector.NewQueryCollector())
	}
}

// Open works like sql.Open with the named driver wrapped.
func Open(driverName, dsn string) (*sql.DB, error) {
	probe, err := sql.Open(driverName, dsn)
	if err != nil {
		return nil, err
	}
	parent := probe.Driver()
	if err := probe.Close(); err != nil {
		return nil, err
	}

	connector, err := (&SQLDriver{parent: parent}).OpenConnector(dsn)
	if err != nil {
		return nil, err
	}

	return sql.OpenDB(connector), nil
}

// Wrap returns d wrapped, for sql.Register.
func Wrap(d driver.Driver) *SQLDriver {
	return &SQLDriver{parent: d}
}

// WrapConnector returns c wrapped, for sql.OpenDB.
func WrapConnector(c driver.Connector) driver.Connector {
	return &sqlConnector{parent: c, driver: &SQLDriver{parent: c.Driver()}}
}

// Open opens a connection through the wrapped driver.
func (d *SQLDriver) Open(name string) (driver.Conn, error) {
	c, err := d.parent.Open(name)
	if err != nil {
		return nil, err
	}

	return &sqlConn{parent: c}, nil
}

// OpenConnector returns a connector through the wrapped driver.
func (d *SQLDriver) OpenConnector(name string) (driver.Connector, error) {
	if dc, ok := d.parent.(driver.DriverContext); ok {
		c, err := dc.OpenConnector(name)
		if err != nil {
			return nil, err
		}
		return &sqlConnector{parent: c, driver: d}, nil
	}

	return &sqlConnector{parent: &dsnConnector{name: name, driver: d.parent}, driver: d}, nil
}

type sqlConnector struct {
	parent driver.Connector
	driver *SQLDriver
}

func (c *sqlConnector) Connect(ctx context.Context) (driver.Conn, error) {
	conn, err := c.parent.Connect(ctx)
	if err != nil {
		return nil, err
	}

	return &sqlConn{parent: conn}, nil
}

func (c *sqlConnector) Driver() driver.Driver {
	return c.driver
}

type dsnConnector struct {
	name   string
	driver driver.Driver
}

func (c *dsnConnector) Connect(context.Context) (driver.Conn, error) {
	return c.driver.Open(c.name)
}

func (c *dsnConnector) Driver() driver.Driver {
	return c.driver
}
