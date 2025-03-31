package services

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"math/rand"
	"time"

	"github.com/MarlonG1/delivery-backend/internal/domain/delivery/constants"
	"github.com/MarlonG1/delivery-backend/internal/domain/delivery/interfaces"
	"github.com/MarlonG1/delivery-backend/internal/domain/delivery/models/entities"
	"github.com/MarlonG1/delivery-backend/internal/domain/delivery/ports"
	"github.com/MarlonG1/delivery-backend/internal/domain/delivery/value_objects"
	errPackage "github.com/MarlonG1/delivery-backend/internal/domain/error"
	"github.com/MarlonG1/delivery-backend/pkg/shared/logs"
)

type OrderService struct {
	repo ports.OrdererRepository
}

func NewOrderService(repo ports.OrdererRepository) interfaces.Orderer {
	return &OrderService{
		repo: repo,
	}
}

func (o OrderService) GetOrderByTrackingNumber(ctx context.Context, trackingNumber string) (*entities.Order, error) {
	order, err := o.repo.GetOrderByTrackingNumber(ctx, trackingNumber)
	if err != nil {
		logs.Error("Failed to get order by tracking number", map[string]interface{}{
			"trackingNumber": trackingNumber,
			"error":          err.Error(),
		})
		return nil, errPackage.NewDomainErrorWithCause("OrderService", "GetOrderByTrackingNumber", "failed to get order by tracking number", err)
	}

	return order, nil
}

func (o OrderService) GetOrdersByClientID(ctx context.Context, clientID string) ([]entities.Order, error) {
	getOrders, err := o.repo.GetOrdersByUserID(ctx, clientID)
	if err != nil {
		logs.Error("Failed to get orders by client id", map[string]interface{}{
			"clientID": clientID,
			"error":    err.Error(),
		})
		return nil, errPackage.NewDomainErrorWithCause("OrderService", "GetOrdersByClientID", "failed to get orders by client id", err)
	}

	return getOrders, nil
}

func (o OrderService) AssignDriverToOrder(ctx context.Context, orderID, driverID string) error {
	err := o.repo.AssignDriverToOrder(ctx, orderID, driverID)
	if err != nil {
		logs.Error("Failed to assign driver to order", map[string]interface{}{
			"orderID":  orderID,
			"driverID": driverID,
			"error":    err.Error(),
		})
		return errPackage.NewDomainErrorWithCause("OrderService", "AssignDriverToOrder", "failed to assign driver to order", err)
	}

	return nil
}

func (o OrderService) CreateOrder(ctx context.Context, order *entities.Order) error {
	if order == nil {
		logs.Error("Order is nil")
		return errPackage.NewDomainError("OrderService", "CreateOrder", "order is nil")
	}

	// 1. Generar estado historico inicial
	statusHistory := &entities.StatusHistory{
		ID:      uuid.NewString(),
		OrderID: order.ID,
		Status:  constants.OrderStatusPending,
	}
	order.StatusHistory = append(order.StatusHistory, *statusHistory)

	// 2. Generar tracking number
	order.TrackingNumber = generateTrackingNumber()

	// 3. Verificar puntos importantes
	if err := order.Validate(); err != nil {
		return err
	}

	//4. Crear pedido
	err := o.repo.CreateOrder(ctx, order)
	if err != nil {
		logs.Error("Failed to create order", map[string]interface{}{
			"orderID":        order.ID,
			"trackingNumber": order.TrackingNumber,
			"error":          err.Error(),
		})
		return errPackage.NewDomainErrorWithCause("OrderService", "CreateOrder", "failed to create order", err)
	}

	//5. Crear QR
	err = o.repo.CreateQRData(ctx, generateQRCode(*order))
	if err != nil {
		logs.Error("Failed to create qr code", map[string]interface{}{
			"orderID":        order.ID,
			"trackingNumber": order.TrackingNumber,
			"error":          err.Error(),
		})
		return errPackage.NewDomainErrorWithCause("OrderService", "CreateOrder", "failed to create qr code", err)
	}

	return nil
}

func (o OrderService) ChangeStatus(ctx context.Context, id, status string) error {
	// 1. Validar que el pedido no este eliminado
	if o.OrderIsDeleted(ctx, id) {
		logs.Warn("Dont change status, order is deleted", map[string]interface{}{
			"orderID": id,
			"error":   errPackage.ErrOrderDeleted.Error(),
		})
		return errPackage.NewDomainErrorWithCause("OrderService", "ChangeStatus", "Dont change status", errPackage.ErrOrderDeleted)
	}

	// 2. Validar que el estado sea valido
	if !value_objects.NewOrderStatus(status).IsValid() {
		logs.Error("Invalid order status", map[string]interface{}{
			"status": status,
		})
		return errPackage.NewDomainError("OrderService", "ChangeStatus", "invalid order status")
	}

	// 3. Obtener pedido para obtener estado actual
	order, err := o.repo.GetOrderByID(ctx, id)
	if err != nil {
		logs.Error("Failed to get order by id", map[string]interface{}{
			"orderID": id,
			"error":   err.Error(),
		})
		return errPackage.NewDomainErrorWithCause("OrderService", "ChangeStatus", "failed to get order by id", err)
	}

	// 4. Validar que la transicion de estados sea valida
	if !value_objects.NewOrderStatus(order.Status).CanTransitionTo(value_objects.NewOrderStatus(status)) {
		logs.Error("Invalid transition", map[string]interface{}{
			"from": order.Status,
			"to":   status,
		})
		return errPackage.NewDomainError("OrderService", "ChangeStatus", fmt.Sprintf("invalid transition from %s to %s", order.Status, status))
	}

	// 5. Cambiar estado
	err = o.repo.ChangeStatus(ctx, id, status)
	if err != nil {
		logs.Error("Failed to change status", map[string]interface{}{
			"orderID": id,
			"status":  status,
			"error":   err.Error(),
		})
		return errPackage.NewDomainErrorWithCause("OrderService", "ChangeStatus", "failed to change status", err)
	}

	return nil
}

func (o OrderService) GetOrderByID(ctx context.Context, orderID string) (*entities.Order, error) {
	order, err := o.repo.GetOrderByID(ctx, orderID)
	if err != nil {
		logs.Error("Failed to get order by id", map[string]interface{}{
			"orderID": orderID,
			"error":   err.Error(),
		})
		return nil, errPackage.NewDomainErrorWithCause("OrderService", "GetOrder", "failed to get order by id", err)
	}

	return order, nil
}

func (o OrderService) GetOrders(ctx context.Context) ([]entities.Order, error) {
	dbOrders, err := o.repo.GetOrders(ctx)
	if err != nil {
		logs.Error("Failed to get orders", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, errPackage.NewDomainErrorWithCause("OrderService", "GetOrders", "failed to get orders", err)
	}

	return dbOrders, nil
}

func (o OrderService) UpdateOrder(ctx context.Context, orderID string, order *entities.Order) error {
	// 1. Verificar si el pedido existe
	dbOrder, err := o.GetOrderByID(ctx, orderID)
	if err != nil {
		logs.Error("Failed to get order by id", map[string]interface{}{
			"orderID": orderID,
			"error":   err.Error(),
		})
		return errPackage.NewDomainErrorWithCause("OrderService", "UpdateOrder", "dont update because the order doesnt exist", err)
	}

	// 2. Verificar si el pedido no esta eliminado
	if dbOrder.DeletedAt != nil {
		logs.Warn("Dont update order, order is deleted", map[string]interface{}{
			"orderID": orderID,
			"error":   errPackage.ErrOrderDeleted.Error(),
		})
		return errPackage.NewDomainErrorWithCause("OrderService", "UpdateOrder", "order is deleted", errPackage.ErrOrderDeleted)
	}

	// 3. Verificar si el pedido esta en un estado que permite actualizacion
	if order.Status != "" && !canUpdateOrder(order) {
		logs.Warn("Dont update order, order is not available for update", map[string]interface{}{
			"orderID": orderID,
			"status":  order.Status,
			"error":   errPackage.ErrCannotUpdateOrder.Error(),
		})
		return errPackage.NewDomainErrorWithCause("OrderService", "UpdateOrder", "order is not available for update", errPackage.ErrCannotUpdateOrder)
	}

	// 4. Actualizar pedido
	err = o.repo.UpdateOrder(ctx, orderID, order)
	if err != nil {
		logs.Error("Failed to update order", map[string]interface{}{
			"orderID": orderID,
			"error":   err.Error(),
		})
		return errPackage.NewDomainErrorWithCause("OrderService", "UpdateOrder", "failed to update order", err)
	}

	return nil
}

func (o OrderService) GetOrdersByCompany(ctx context.Context, companyID string, params *entities.OrderQueryParams) ([]entities.Order, int64, error) {
	orders, total, err := o.repo.GetOrdersByCompany(ctx, companyID, params)
	if err != nil {
		logs.Error("Failed to get orders by company", map[string]interface{}{
			"companyID": companyID,
			"error":     err.Error(),
		})
		return nil, 0, errPackage.NewDomainErrorWithCause("OrderService", "GetOrdersByCompany", "failed to get orders by company", err)
	}

	return orders, total, nil
}

func (o OrderService) SoftDeleteOrder(ctx context.Context, id string) error {
	// 1. Verificar si el pedido no esta eliminado
	if o.OrderIsDeleted(ctx, id) {
		logs.Warn("Dont delete order, order is already deleted", map[string]interface{}{
			"orderID": id,
			"error":   errPackage.ErrOrderAlreadyDeleted.Error(),
		})
		return errPackage.NewDomainErrorWithCause("OrderService", "SoftDeleteOrder", "order is already deleted", errPackage.ErrOrderAlreadyDeleted)
	}

	// 2. Verificar si el pedido esta en un estado que permite eliminacion
	err := o.repo.SoftDeleteOrder(ctx, id)
	if err != nil {
		logs.Error("Failed to soft delete order", map[string]interface{}{
			"orderID": id,
			"error":   err.Error(),
		})
		return errPackage.NewDomainErrorWithCause("OrderService", "SoftDeleteOrder", "failed to soft delete order", err)
	}

	return nil
}

func (o OrderService) RestoreOrder(ctx context.Context, id string) error {
	// 1. Verificar si el pedido esta eliminado
	if !o.OrderIsDeleted(ctx, id) {
		logs.Warn("Dont restore order, order is not deleted", map[string]interface{}{
			"orderID": id,
			"error":   errPackage.ErrOrderNotDeleted.Error(),
		})
		return errPackage.NewDomainErrorWithCause("OrderService", "RestoreOrder", "order is not deleted", errPackage.ErrOrderNotDeleted)
	}

	err := o.repo.RestoreOrder(ctx, id)
	if err != nil {
		logs.Error("Failed to restore order", map[string]interface{}{
			"orderID": id,
			"error":   err.Error(),
		})
		return errPackage.NewDomainErrorWithCause("OrderService", "RestoreOrder", "failed to restore order", err)
	}

	return nil
}

func (o OrderService) OrderIsDeleted(ctx context.Context, orderID string) bool {
	order, err := o.GetOrderByID(ctx, orderID)
	if err != nil {
		logs.Error("Failed to get order by id in OrderIsDeleted method", map[string]interface{}{
			"orderID": orderID,
			"error":   err.Error(),
		})
		return false
	}

	return order.DeletedAt != nil
}

func (o OrderService) IsAvailableForDelete(ctx context.Context, orderID string) error {
	order, err := o.GetOrderByID(ctx, orderID)
	if err != nil {
		logs.Error("Failed to get order by id in start of IsAvailableForDelete", map[string]interface{}{
			"orderID": orderID,
			"error":   err.Error(),
		})
		return err
	}

	if order.DeletedAt != nil {
		logs.Warn("Order is already deleted", map[string]interface{}{
			"orderID":   orderID,
			"error":     errPackage.ErrOrderAlreadyDeleted.Error(),
			"deletedAt": order.DeletedAt,
		})
		return errPackage.NewDomainErrorWithCause("OrderService", "IsAvailableForDelete", "an error occurred", errPackage.ErrOrderAlreadyDeleted)
	}

	if !canDeleteOrder(order) {
		logs.Warn("Order is not available for delete", map[string]interface{}{
			"orderID": orderID,
			"error":   errPackage.ErrCannotDeleteOrder.Error(),
			"status":  order.Status,
		})
		return errPackage.NewDomainErrorWithCause("OrderService", "IsAvailableForDelete", "order is not available for delete", errPackage.ErrCannotDeleteOrder)
	}

	return nil
}

func generateQRCode(order entities.Order) *entities.QRCode {
	return &entities.QRCode{
		OrderID: order.ID,
		QRData:  order.TrackingNumber,
	}
}

func canDeleteOrder(order *entities.Order) bool {
	return constants.AllowedStatesToDelete[order.Status]
}

func canUpdateOrder(order *entities.Order) bool {
	return constants.AllowedStatesToUpdate[order.Status]
}

func generateTrackingNumber() string {
	// Formato: [prefijo]-[timestamp]-[aleatorio]
	prefix := "DEL"
	timestamp := time.Now().Format("060102")
	random := fmt.Sprintf("%04d", rand.Intn(10000))

	return fmt.Sprintf("%s-%s-%s", prefix, timestamp, random)
}
