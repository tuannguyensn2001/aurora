package router

import (
	"net/http"
	"strconv"

	"api/config"
	"api/internal/dto"
	"api/internal/handler"
	"api/internal/middleware"
	"api/internal/model"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

// Router handles HTTP routing using gin
type Router struct {
	handler *handler.Handler
	logger  zerolog.Logger
	config  *config.Config
}

// New creates a new router instance
func New(h *handler.Handler, logger zerolog.Logger, cfg *config.Config) *Router {
	return &Router{
		handler: h,
		logger:  logger,
		config:  cfg,
	}
}

// SetupRoutes configures all routes
func (r *Router) SetupRoutes(engine *gin.Engine) {
	// Add logger middleware
	engine.Use(r.loggingMiddleware())

	// Add error handling middleware
	engine.Use(r.errorHandlingMiddleware())

	// Health check
	engine.GET("/health", r.healthCheck)

	// Auth routes (no versioning for OAuth)

	// API v1 routes
	v1 := engine.Group("/api/v1")
	{
		// Auth routes (public - no JWT middleware)
		auth := v1.Group("/auth")
		{
			auth.GET("/google/login", r.googleLogin)
			auth.POST("/google/callback", r.googleCallback)
			auth.POST("/refresh", r.refreshToken)

			// Protected auth routes (require JWT)
			authProtected := auth.Group("")
			authProtected.Use(middleware.JWTMiddleware(r.config))
			{
				authProtected.GET("/me", r.getCurrentUser)
			}
		}

		// SDK routes (public - no JWT middleware, accessed by client SDKs)
		sdk := v1.Group("/sdk")
		{
			sdk.POST("/metadata", r.getMetadataSDK)
			sdk.POST("/parameters", r.getAllParametersSDK)
			sdk.POST("/experiments", r.getAllExperimentsSDK)
		}

		// Protected routes group (require JWT authentication)
		protected := v1.Group("")
		protected.Use(middleware.JWTMiddleware(r.config))
		{
			// Attribute routes
			attributes := protected.Group("/attributes")
			{
				attributes.POST("", r.createAttribute)
				attributes.GET("", r.getAllAttributes)
				attributes.GET("/:id", r.getAttributeByID)
				attributes.PATCH("/:id", r.updateAttribute)
				attributes.DELETE("/:id", r.deleteAttribute)
				attributes.PATCH("/:id/increment-usage", r.incrementAttributeUsageCount)
				attributes.PATCH("/:id/decrement-usage", r.decrementAttributeUsageCount)
			}

			// Segment routes
			segments := protected.Group("/segments")
			{
				segments.POST("", r.createSegment)
				segments.GET("", r.getAllSegments)
				segments.GET("/:id", r.getSegmentByID)
				segments.PATCH("/:id", r.updateSegment)
				segments.DELETE("/:id", r.deleteSegment)
				segments.POST("/check-overlap", r.checkSegmentOverlap)
			}

			// Parameter routes
			parameters := protected.Group("/parameters")
			{
				parameters.POST("", r.createParameter)
				parameters.GET("", r.getAllParameters)
				parameters.GET("/:id", r.getParameterByID)
				parameters.PATCH("/:id", r.updateParameter)
				parameters.PUT("/:id", r.updateParameterWithRules)
				parameters.DELETE("/:id", r.deleteParameter)
				parameters.POST("/simulate", r.simulateParameter)

				// Parameter change request routes
				parameters.POST("/:id/change-requests", r.createParameterChangeRequest)
				parameters.GET("/:id/change-requests/pending", r.getPendingParameterChangeRequest)
			}

			// Parameter Change Request routes
			changeRequests := protected.Group("/parameter-change-requests")
			{
				changeRequests.GET("", r.getParameterChangeRequestsByStatus)
				changeRequests.GET("/:id", r.getParameterChangeRequestByID)
				changeRequests.GET("/:id/details", r.getParameterChangeRequestByIDWithDetails)
				changeRequests.PATCH("/:id/approve", r.approveParameterChangeRequest)
				changeRequests.PATCH("/:id/reject", r.rejectParameterChangeRequest)
			}

			// Experiment routes
			experiments := protected.Group("/experiments")
			{
				experiments.POST("", r.createExperiment)
				experiments.GET("", r.getAllExperiments)
				experiments.GET("/:id", r.getExperimentByID)
				experiments.PATCH("/:id/reject", r.rejectExperiment)
				experiments.PATCH("/:id/approve", r.approveExperiment)
				experiments.PATCH("/:id/abort", r.abortExperiment)
			}
		}
	}
}

// Middleware functions
func (r *Router) loggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Add logger to context
		ctx := r.logger.WithContext(c.Request.Context())
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

func (r *Router) errorHandlingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Handle any errors that occurred during request processing
		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err
			r.handleError(c, err)
		}
	}
}

// Error handling
func (r *Router) handleError(c *gin.Context, err error) {
	statusCode := http.StatusInternalServerError
	errorType := "Internal Server Error"

	errMsg := err.Error()
	if contains(errMsg, "not found") {
		statusCode = http.StatusNotFound
		errorType = "Not Found"
	} else if contains(errMsg, "already exists") || contains(errMsg, "required") || contains(errMsg, "cannot delete") {
		statusCode = http.StatusConflict
		errorType = "Conflict"
	} else if contains(errMsg, "invalid") || contains(errMsg, "bad request") {
		statusCode = http.StatusBadRequest
		errorType = "Bad Request"
	}

	c.JSON(statusCode, dto.ErrorResponse{
		Error:   errorType,
		Message: errMsg,
	})
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && containsHelper(s, substr)))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Helper function to parse ID from URL parameter
func parseIDParam(c *gin.Context) (uint, error) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}

// Health check handler
func (r *Router) healthCheck(c *gin.Context) {
	result, err := r.handler.HealthCheck(c.Request.Context())
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": result})
}

// Attribute handlers
func (r *Router) createAttribute(c *gin.Context) {
	var req dto.CreateAttributeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}

	result, err := r.handler.CreateAttribute(c.Request.Context(), &req)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, result)
}

func (r *Router) getAllAttributes(c *gin.Context) {
	result, err := r.handler.GetAllAttributes(c.Request.Context())
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, result)
}

func (r *Router) getAttributeByID(c *gin.Context) {
	id, err := parseIDParam(c)
	if err != nil {
		c.Error(err)
		return
	}

	result, err := r.handler.GetAttributeByID(c.Request.Context(), id)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, result)
}

func (r *Router) updateAttribute(c *gin.Context) {
	id, err := parseIDParam(c)
	if err != nil {
		c.Error(err)
		return
	}

	var req dto.UpdateAttributeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}

	result, err := r.handler.UpdateAttribute(c.Request.Context(), id, &req)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, result)
}

func (r *Router) deleteAttribute(c *gin.Context) {
	id, err := parseIDParam(c)
	if err != nil {
		c.Error(err)
		return
	}

	err = r.handler.DeleteAttribute(c.Request.Context(), id)
	if err != nil {
		c.Error(err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (r *Router) incrementAttributeUsageCount(c *gin.Context) {
	id, err := parseIDParam(c)
	if err != nil {
		c.Error(err)
		return
	}

	err = r.handler.IncrementAttributeUsageCount(c.Request.Context(), id)
	if err != nil {
		c.Error(err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (r *Router) decrementAttributeUsageCount(c *gin.Context) {
	id, err := parseIDParam(c)
	if err != nil {
		c.Error(err)
		return
	}

	err = r.handler.DecrementAttributeUsageCount(c.Request.Context(), id)
	if err != nil {
		c.Error(err)
		return
	}

	c.Status(http.StatusNoContent)
}

// Segment handlers
func (r *Router) createSegment(c *gin.Context) {
	var req dto.CreateSegmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}

	result, err := r.handler.CreateSegment(c.Request.Context(), &req)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, result)
}

func (r *Router) getAllSegments(c *gin.Context) {
	result, err := r.handler.GetAllSegments(c.Request.Context())
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, result)
}

func (r *Router) getSegmentByID(c *gin.Context) {
	id, err := parseIDParam(c)
	if err != nil {
		c.Error(err)
		return
	}

	result, err := r.handler.GetSegmentByID(c.Request.Context(), id)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, result)
}

func (r *Router) updateSegment(c *gin.Context) {
	id, err := parseIDParam(c)
	if err != nil {
		c.Error(err)
		return
	}

	var req dto.UpdateSegmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}

	result, err := r.handler.UpdateSegment(c.Request.Context(), id, &req)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, result)
}

func (r *Router) deleteSegment(c *gin.Context) {
	id, err := parseIDParam(c)
	if err != nil {
		c.Error(err)
		return
	}

	err = r.handler.DeleteSegment(c.Request.Context(), id)
	if err != nil {
		c.Error(err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (r *Router) checkSegmentOverlap(c *gin.Context) {
	var req dto.CheckSegmentOverlapRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}

	result, err := r.handler.CheckSegmentOverlap(c.Request.Context(), &req)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, result)
}

// Parameter handlers
func (r *Router) createParameter(c *gin.Context) {
	var req dto.CreateParameterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}

	result, err := r.handler.CreateParameter(c.Request.Context(), &req)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, result)
}

func (r *Router) getAllParameters(c *gin.Context) {
	result, err := r.handler.GetAllParameters(c.Request.Context())
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, result)
}

func (r *Router) getParameterByID(c *gin.Context) {
	id, err := parseIDParam(c)
	if err != nil {
		c.Error(err)
		return
	}

	result, err := r.handler.GetParameterByID(c.Request.Context(), id)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, result)
}

func (r *Router) updateParameter(c *gin.Context) {
	id, err := parseIDParam(c)
	if err != nil {
		c.Error(err)
		return
	}

	var req dto.UpdateParameterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}

	result, err := r.handler.UpdateParameter(c.Request.Context(), id, &req)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, result)
}

func (r *Router) updateParameterWithRules(c *gin.Context) {
	id, err := parseIDParam(c)
	if err != nil {
		c.Error(err)
		return
	}

	var req dto.UpdateParameterWithRulesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}

	result, err := r.handler.UpdateParameterWithRules(c.Request.Context(), id, &req)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, result)
}

func (r *Router) deleteParameter(c *gin.Context) {
	id, err := parseIDParam(c)
	if err != nil {
		c.Error(err)
		return
	}

	err = r.handler.DeleteParameter(c.Request.Context(), id)
	if err != nil {
		c.Error(err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (r *Router) simulateParameter(c *gin.Context) {
	var req dto.SimulateParameterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}

	result, err := r.handler.SimulateParameter(c.Request.Context(), &req)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, result)
}

// Experiment handlers
func (r *Router) createExperiment(c *gin.Context) {
	var req dto.CreateExperimentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}

	result, err := r.handler.CreateExperiment(c.Request.Context(), &req)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, result)
}

func (r *Router) getAllExperiments(c *gin.Context) {
	result, err := r.handler.GetAllExperiments(c.Request.Context())
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, result)
}

func (r *Router) getExperimentByID(c *gin.Context) {
	id, err := parseIDParam(c)
	if err != nil {
		c.Error(err)
		return
	}

	result, err := r.handler.GetExperimentByID(c.Request.Context(), id)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, result)
}

func (r *Router) rejectExperiment(c *gin.Context) {
	id, err := parseIDParam(c)
	if err != nil {
		c.Error(err)
		return
	}

	var req dto.RejectExperimentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}

	result, err := r.handler.RejectExperiment(c.Request.Context(), id, &req)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, result)
}

func (r *Router) approveExperiment(c *gin.Context) {
	id, err := parseIDParam(c)
	if err != nil {
		c.Error(err)
		return
	}

	var req dto.ApproveExperimentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}

	result, err := r.handler.ApproveExperiment(c.Request.Context(), id, &req)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, result)
}

func (r *Router) abortExperiment(c *gin.Context) {
	id, err := parseIDParam(c)
	if err != nil {
		c.Error(err)
		return
	}

	var req dto.AbortExperimentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}

	result, err := r.handler.AbortExperiment(c.Request.Context(), id, &req)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, result)
}

// SDK handlers
func (r *Router) getMetadataSDK(c *gin.Context) {
	var req dto.GetMetadataSDKRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}

	result, err := r.handler.GetMetdataSDK(c.Request.Context(), &req)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, result)
}

func (r *Router) getAllParametersSDK(c *gin.Context) {

	var req dto.GetAllParametersSDKRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}

	result, err := r.handler.GetAllParametersSDK(c.Request.Context(), &req)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, result)

}

func (r *Router) getAllExperimentsSDK(c *gin.Context) {
	var req dto.GetAllExperimentsSDKRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}

	result, err := r.handler.GetAllExperimentsSDK(c.Request.Context(), &req)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, result)
}

// Auth handlers
func (r *Router) googleLogin(c *gin.Context) {
	result, err := r.handler.GoogleLogin(c.Request.Context())
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, result)
}

func (r *Router) googleCallback(c *gin.Context) {
	var req dto.GoogleCallbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}

	result, err := r.handler.GoogleCallback(c.Request.Context(), &req)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, result)
}

func (r *Router) refreshToken(c *gin.Context) {
	var req dto.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}

	result, err := r.handler.RefreshToken(c.Request.Context(), &req)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, result)
}

func (r *Router) getCurrentUser(c *gin.Context) {
	// Extract user ID from JWT token (set by JWT middleware)
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	result, err := r.handler.GetCurrentUser(c.Request.Context(), userID)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, result)
}

// Parameter Change Request handlers
func (r *Router) createParameterChangeRequest(c *gin.Context) {
	id, err := parseIDParam(c)
	if err != nil {
		c.Error(err)
		return
	}

	// Extract user ID from JWT token
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req dto.CreateParameterChangeRequestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}

	// Set parameter ID from URL
	req.ParameterID = id

	result, err := r.handler.CreateParameterChangeRequest(c.Request.Context(), userID, &req)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, result)
}

func (r *Router) getParameterChangeRequestByID(c *gin.Context) {
	id, err := parseIDParam(c)
	if err != nil {
		c.Error(err)
		return
	}

	result, err := r.handler.GetParameterChangeRequestByID(c.Request.Context(), id)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, result)
}

func (r *Router) getPendingParameterChangeRequest(c *gin.Context) {
	id, err := parseIDParam(c)
	if err != nil {
		c.Error(err)
		return
	}

	result, err := r.handler.GetPendingParameterChangeRequestByParameterID(c.Request.Context(), id)
	if err != nil {
		c.Error(err)
		return
	}

	if result == nil {
		c.JSON(http.StatusOK, nil)
		return
	}

	c.JSON(http.StatusOK, result)
}

func (r *Router) getParameterChangeRequests(c *gin.Context) {
	id, err := parseIDParam(c)
	if err != nil {
		c.Error(err)
		return
	}

	result, err := r.handler.GetParameterChangeRequestsByParameterID(c.Request.Context(), id)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, result)
}

func (r *Router) approveParameterChangeRequest(c *gin.Context) {
	id, err := parseIDParam(c)
	if err != nil {
		c.Error(err)
		return
	}

	// Extract user ID from JWT token
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req dto.ApproveParameterChangeRequestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}

	result, err := r.handler.ApproveParameterChangeRequest(c.Request.Context(), id, userID, &req)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, result)
}

func (r *Router) rejectParameterChangeRequest(c *gin.Context) {
	id, err := parseIDParam(c)
	if err != nil {
		c.Error(err)
		return
	}

	// Extract user ID from JWT token
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req dto.RejectParameterChangeRequestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}

	result, err := r.handler.RejectParameterChangeRequest(c.Request.Context(), id, userID, &req)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, result)
}

func (r *Router) getParameterChangeRequestsByStatus(c *gin.Context) {
	// Parse query parameters
	status := c.Query("status")
	if status == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "status parameter is required"})
		return
	}

	// Validate status
	var changeRequestStatus model.ParameterChangeRequestStatus
	switch status {
	case "pending":
		changeRequestStatus = model.ChangeRequestStatusPending
	case "approved":
		changeRequestStatus = model.ChangeRequestStatusApproved
	case "rejected":
		changeRequestStatus = model.ChangeRequestStatusRejected
	case "cancelled":
		changeRequestStatus = model.ChangeRequestStatusCancelled
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid status. Must be one of: pending, approved, rejected, cancelled"})
		return
	}

	// Parse pagination parameters
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid limit parameter"})
		return
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid offset parameter"})
		return
	}

	result, err := r.handler.GetParameterChangeRequestsByStatus(c.Request.Context(), changeRequestStatus, limit, offset)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, result)
}

func (r *Router) getParameterChangeRequestByIDWithDetails(c *gin.Context) {
	id, err := parseIDParam(c)
	if err != nil {
		c.Error(err)
		return
	}

	result, err := r.handler.GetParameterChangeRequestByIDWithDetails(c.Request.Context(), id)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, result)
}
