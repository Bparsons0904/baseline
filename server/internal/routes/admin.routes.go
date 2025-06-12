package routes

import (
	"server/internal/app"
	adminController "server/internal/controllers/admin"
	"server/internal/logger"
	. "server/internal/models"

	"github.com/gofiber/fiber/v2"
)

type AdminRoute struct {
	Route
	controller adminController.AdminController
}

func NewAdminRoute(app app.App, router fiber.Router) *AdminRoute {
	log := logger.New("routes").File("admin.routes")
	
	if app.AdminController == nil {
		log.Warn("AdminController is nil in app")
		return &AdminRoute{
			controller: adminController.AdminController{},
			Route: Route{
				log:        log,
				router:     router,
				middleware: app.Middleware,
			},
		}
	}
	
	return &AdminRoute{
		controller: *app.AdminController,
		Route: Route{
			log:        log,
			router:     router,
			middleware: app.Middleware,
		},
	}
}

func (r *AdminRoute) Register() {
	users := r.router.Group("/admin")
	users.Post("/broadcast", r.broadcast)
}

func (r *AdminRoute) broadcast(c *fiber.Ctx) error {
	log := r.log.Function("broadcast")
	log.Info("Broadcasting admin message")

	type Response struct {
		Message string `json:"message"`
	}

	var response Response
	if err := c.BodyParser(&response); err != nil {
		log.Er("failed to parse login request", err)
		return c.Status(fiber.StatusBadRequest).
			JSON(fiber.Map{"message": "failed to parse login request"})
	}

	user := c.Locals("user").(User)
	if user.ID == "" {
		log.ErMsg("No user found in locals")
		return c.Status(fiber.StatusInternalServerError).
			JSON(fiber.Map{"message": "failed to get user"})
	}

	r.controller.SendBroadcast(c.Context(), user, response.Message)

	return c.JSON(fiber.Map{"message": "Broadcast sent"})
}
