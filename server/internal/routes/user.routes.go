package routes

import (
	"server/internal/app"
	userController "server/internal/controllers/users"
	"server/internal/logger"
	. "server/internal/models"
	"server/internal/utils"

	"github.com/gofiber/fiber/v2"
)

type UserRoute struct {
	Route
	controller userController.UserController
}

func NewUserRoute(app app.App, router fiber.Router) *UserRoute {
	log := logger.New("routes").File("user.routes")
	
	if app.UserController == nil {
		log.Warn("UserController is nil in app")
		return &UserRoute{
			controller: userController.UserController{},
			Route: Route{
				log:        log,
				router:     router,
				middleware: app.Middleware,
			},
		}
	}
	
	return &UserRoute{
		controller: *app.UserController,
		Route: Route{
			log:        log,
			router:     router,
			middleware: app.Middleware,
		},
	}
}

func (r *UserRoute) Register() {
	users := r.router.Group("/users")
	users.Post("/login", r.login)

	users.Use(r.middleware.BasicAuth(), r.middleware.AuthNoContent())
	users.Get("/", r.getUser)
	users.Post("/logout", r.logout)
}

func (r *UserRoute) getUser(c *fiber.Ctx) error {
	user := c.Locals("user").(User)
	session := c.Locals("session").(Session)
	if user.ID == "" {
		r.log.Function("getUser").ErMsg("No user found in locals")
		return c.Status(fiber.StatusInternalServerError).
			JSON(fiber.Map{"message": "failed to get user"})
	}

	utils.ApplyToken(c, session.Token) // TODO: Why is this needed? Wouldn't the middleware do this?

	return c.JSON(fiber.Map{"message": "User logged in", "user": user})
}

func (r *UserRoute) logout(c *fiber.Ctx) error {
	log := r.log.Function("logout")
	sessionID := c.Cookies(SESSION_COOKIE_KEY)

	utils.ExpireCookie(c, SESSION_COOKIE_KEY)

	err := r.controller.Logout(sessionID)
	if err != nil {
		log.Er("failed to logout", err)
		return c.Status(fiber.StatusInternalServerError).
			JSON(fiber.Map{"message": "failed to logout"})

	}

	return c.JSON(fiber.Map{"message": "User logged out"})
}

func (r *UserRoute) login(c *fiber.Ctx) error {
	log := r.log.Function("login")

	var loginRequest LoginRequest
	if err := c.BodyParser(&loginRequest); err != nil {
		log.Er("failed to parse login request", err)
		return c.Status(fiber.StatusBadRequest).
			JSON(fiber.Map{"message": "failed to parse login request"})
	}

	user, session, err := r.controller.Login(c.Context(), loginRequest)
	if err != nil {
		log.Er("failed to login", err)
		return c.Status(fiber.StatusInternalServerError).
			JSON(fiber.Map{"message": "Failed to login"})
	}

	applySessionResponse(c, session)

	return c.JSON(fiber.Map{"message": "User logged in", "user": user})
}

func applySessionResponse(c *fiber.Ctx, session Session) {
	utils.ApplyCookie(c, utils.Cookie{
		Name:    SESSION_COOKIE_KEY,
		Value:   session.ID,
		Expires: session.ExpiresAt,
	})

	utils.ApplyToken(c, session.Token)
}
