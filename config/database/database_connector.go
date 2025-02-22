package database

import (
	errPackage "config/error"
	"fmt"
	"gorm.io/gorm"
	"shared/logs"
)

type DriverConfig interface {
	GetDSN() gorm.Dialector
	GetDriverName() string
	GetHost() string
}

type DbConnection struct {
	Db     *gorm.DB
	Config *gorm.Config
	Driver DriverConfig
	Err    error
}

func NewDatabaseConnection(driver DriverConfig) *DbConnection {
	return &DbConnection{
		Driver: driver,
		Config: &gorm.Config{},
	}
}

func (d *DbConnection) Open() error {
	d.Db, d.Err = gorm.Open(d.Driver.GetDSN(), d.Config)
	if d.Err != nil {
		logs.Error(errPackage.ErrFailedToConnectDb.Error(), map[string]interface{}{
			"Database type:": d.Driver.GetDriverName(),
			"Database error": d.Err.Error(),
		})
		return d.Err
	}

	logs.Info("Database connection has been set successfully", map[string]interface{}{
		"Database type:": d.Driver.GetDriverName(),
		"Database host":  d.Driver.GetHost(),
	})

	return nil
}

func (d *DbConnection) Close() error {
	dbInstance, err := d.Db.DB()
	if err != nil {
		logs.Error(errPackage.ErrFailedToGetDBInstance.Error(), map[string]interface{}{
			"Database type:": d.Driver.GetDriverName(),
			"Database error": err.Error(),
		})
		return err
	}

	if err := dbInstance.Close(); err != nil {
		logs.Error(errPackage.ErrFailedToCloseDbConnection.Error(), map[string]interface{}{
			"Database type:": d.Driver.GetDriverName(),
			"Database error": err.Error(),
		})
		return err
	}

	logs.Info(fmt.Sprintf("Database connection close successfully"), map[string]interface{}{
		"Database type:": d.Driver.GetDriverName(),
	})

	return nil
}
