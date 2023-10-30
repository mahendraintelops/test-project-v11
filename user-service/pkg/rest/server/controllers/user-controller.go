package controllers

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/mahendraintelops/test-project-v11/user-service/pkg/rest/server/daos/clients/sqls"
	"github.com/mahendraintelops/test-project-v11/user-service/pkg/rest/server/models"
	"github.com/mahendraintelops/test-project-v11/user-service/pkg/rest/server/services"
	log "github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"net/http"
	"os"
	"strconv"
)

type UserController struct {
	userService *services.UserService
}

func NewUserController() (*UserController, error) {
	userService, err := services.NewUserService()
	if err != nil {
		return nil, err
	}
	return &UserController{
		userService: userService,
	}, nil
}

func (userController *UserController) CreateUser(context *gin.Context) {
	// validate input
	var input models.User
	if err := context.ShouldBindJSON(&input); err != nil {
		log.Error(err)
		context.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	// trigger user creation
	userCreated, err := userController.userService.CreateUser(&input)
	if err != nil {
		log.Error(err)
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	context.JSON(http.StatusCreated, userCreated)
}

func (userController *UserController) ListUsers(context *gin.Context) {
	// trigger all users fetching
	users, err := userController.userService.ListUsers()
	if err != nil {
		log.Error(err)
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	context.JSON(http.StatusOK, users)
}

func (userController *UserController) FetchUser(context *gin.Context) {
	id, err := strconv.ParseInt(context.Param("id"), 10, 64)
	if err != nil {
		log.Error(err)
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// trigger user fetching
	user, err := userController.userService.GetUser(id)
	if err != nil {
		log.Error(err)
		if errors.Is(err, sqls.ErrNotExists) {
			context.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	serviceName := os.Getenv("SERVICE_NAME")
	collectorURL := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if len(serviceName) > 0 && len(collectorURL) > 0 {
		// get the current span by the request context
		currentSpan := trace.SpanFromContext(context.Request.Context())
		currentSpan.SetAttributes(attribute.String("user.id", strconv.FormatInt(user.Id, 10)))
	}

	context.JSON(http.StatusOK, user)
}
