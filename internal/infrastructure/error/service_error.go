package error

import (
	"errors"
	"fmt"
	"github.com/MarlonG1/delivery-backend/configs"
	"gorm.io/gorm"
)

type ServiceError struct {
	Type      string
	Operation string
	Err       error
}

func (e *ServiceError) Error() string {
	envConfig, err := config.NewEnvConfig()
	if err != nil {
		return ""
	}
	if envConfig.Server.Debug {
		return fmt.Sprintf("[%s] %s: | cause: %v", e.Type, e.Operation, e.Err.Error())
	}

	return fmt.Sprintf("%s", e.Err.Error())
}

// NewGeneralServiceError crea un nuevo error de servicio general con el tipo de servicio, la operación, el mensaje y el error.
func NewGeneralServiceError(serviceType, op string, err error) *ServiceError {
	err = IsGormError(err)
	return &ServiceError{
		Type:      serviceType,
		Operation: op,
		Err:       err,
	}
}

func IsGormError(err error) error {
	if err == nil {
		return nil
	}

	gormErrors := []error{
		gorm.ErrDuplicatedKey,
		gorm.ErrRecordNotFound,
		gorm.ErrForeignKeyViolated,
		gorm.ErrPrimaryKeyRequired,
		gorm.ErrInvalidData,
	}

	for _, gormErr := range gormErrors {
		if errors.Is(err, gormErr) {
			return ErrGenericDBError
		}
	}

	return err
}
