package bootstrap

import "infrastructure/api/handlers"

type HandlerContainer struct {
	usesCases *UseCaseContainer

	authHandler *handlers.AuthHandler
	userHandler *handlers.UserHandler
}

func NewHandlerContainer(userCases *UseCaseContainer) *HandlerContainer {
	return &HandlerContainer{
		usesCases: userCases,
	}
}

func (c *HandlerContainer) Initialize() error {
	c.authHandler = handlers.NewAuthHandler(c.usesCases.GetAuthUseCase())
	c.userHandler = handlers.NewUserHandler(c.usesCases.GetUserUseCase())

	return nil
}

func (c *HandlerContainer) GetAuthHandler() *handlers.AuthHandler {
	return c.authHandler
}

func (c *HandlerContainer) GetUserHandler() *handlers.UserHandler {
	return c.userHandler
}
