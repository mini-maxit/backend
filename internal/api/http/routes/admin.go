package routes

import (
	"github.com/gorilla/mux"
	"github.com/mini-maxit/backend/package/utils"
	"go.uber.org/zap"
)

// AdminRoute defines all admin-specific routes with full access.
type AdminRoute interface {
	Placeholder() // Placeholder method to ensure interface is not empty
}

type adminRouteImpl struct {
	logger *zap.SugaredLogger
}

func (ar *adminRouteImpl) Placeholder() {}

// NewAdminRoute creates a new AdminRoute.
func NewAdminRoute() AdminRoute {
	return &adminRouteImpl{
		logger: utils.NewNamedLogger("admin-routes"),
	}
}

// RegisterAdminRoutes registers all admin routes.
func RegisterAdminRoutes(mux *mux.Router, route AdminRoute) {
	// Task routes
}
