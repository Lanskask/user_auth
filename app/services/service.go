package services

import (
	"configs"
	"fmt"
	"handlers"
	"model"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	db    model.IDB
	store *session.Store
	conf  configs.Config
}

type user struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func NewService(db model.IDB, store *session.Store, conf configs.Config) *Service {
	return &Service{
		store: store,
		db:    db,
		conf:  conf,
	}
}

type IService interface {
	Register(c *fiber.Ctx) error
	Login(c *fiber.Ctx) error
	Logout(c *fiber.Ctx) error
	HealthCheck(c *fiber.Ctx) error
	GetUser(c *fiber.Ctx) error
}

func (s *Service) Register(c *fiber.Ctx) error {
	c.Accepts(fiber.MIMEApplicationJSON)
	var data user

	err := c.BodyParser(&data)
	if err != nil {
		return handlers.HandleInternalSerErr(c, err, "error parsing register post body")
	}

	password, err := bcrypt.GenerateFromPassword([]byte(data.Password), s.conf.BcryptCost)
	if err != nil {
		return handlers.HandleInternalSerErr(c, err, "error generating bcrypt hash")
	}

	user := model.User{
		Name:     data.Name,
		Email:    data.Email,
		Password: string(password),
	}

	if _, err := s.db.CreateUser(&user); err != nil {
		return handlers.HandleInternalSerErr(c, err, "error creating user in the DB")
	}

	return handlers.HandleOk(c, "registered")
}

func (s *Service) Login(c *fiber.Ctx) error {
	var data user
	if err := c.BodyParser(&data); err != nil {
		return handlers.HandleInternalSerErr(c, err, "error generating bcrypt hash")
	}
	var user model.User
	exists, err := s.db.CheckEmailExistence(data.Email, &user)
	if err != nil {
		return handlers.HandleInternalSerErr(c, err, "error check email existence")
	}
	if !exists {
		return handlers.HandleUnAuth(c)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(data.Password)); err != nil {
		return handlers.HandleUnAuth(c)
	}

	sess, err := s.store.Get(c)
	if err != nil {
		return handlers.HandleInternalSerErr(c, err, "error getting session from context for this user")
	}

	sess.Set(s.conf.AuthKey, true)
	sess.Set(s.conf.UserIDKey, user.ID)
	if err := sess.Save(); err != nil {
		return handlers.HandleInternalSerErr(c, err, "error save session for this user")
	}

	return handlers.HandleOk(c, "login successfully")
}

func (s *Service) Logout(c *fiber.Ctx) error {
	sess, err := s.store.Get(c)
	if err != nil {
		return handlers.HandleOk(c, "logged out (session not found)")
	}

	if err := sess.Destroy(); err != nil {
		return handlers.HandleInternalSerErr(c, err, "failed to destroy a session")
	}

	return handlers.HandleOk(c, "logged out")
}

func (s *Service) HealthCheck(c *fiber.Ctx) error {
	sess, err := s.store.Get(c)
	if err != nil {
		return handlers.HandleOk(c, "not authorised")
	}

	if auth := sess.Get(s.conf.AuthKey); auth != nil {
		return handlers.HandleOk(c, "authenticated")
	}

	return handlers.HandleOk(c, "not authorised")
}

func (s *Service) GetUser(c *fiber.Ctx) error {
	sess, err := s.store.Get(c)
	if err != nil {
		return handlers.HandleInternalSerErr(c, err, "failed to get session from store")
	}

	if sess.Get(s.conf.AuthKey) == nil {
		return handlers.HandleUnAuth(c)
	}

	userID := sess.Get(s.conf.UserIDKey)
	if userID == nil {
		return handlers.HandleUnAuth(c)
	}

	var user model.User
	//user, err = s.db.GetUser(strconv.Itoa(userID.(int)))
	user, err = s.db.GetUser(fmt.Sprint(userID))
	if err != nil {
		return handlers.HandleUnAuth(c)
	}

	user.Password = ""
	return c.Status(fiber.StatusOK).JSON(user)
}
