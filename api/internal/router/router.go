package router

import (
	"net/http"
	"strconv"

	"api/internal/dto"
	"api/internal/handler"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

// Router handles HTTP routing using gin
type Router struct {
	handler *handler.Handler
	logger  zerolog.Logger
}

// New creates a new router instance
func New(h *handler.Handler, logger zerolog.Logger) *Router {
	return &Router{
		handler: h,
		logger:  logger,
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

	// API v1 routes
	v1 := engine.Group("/api/v1")
	{
		// Attribute routes
		attributes := v1.Group("/attributes")
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
		segments := v1.Group("/segments")
		{
			segments.POST("", r.createSegment)
			segments.GET("", r.getAllSegments)
			segments.GET("/:id", r.getSegmentByID)
			segments.PATCH("/:id", r.updateSegment)
			segments.DELETE("/:id", r.deleteSegment)
		}

		// Parameter routes
		parameters := v1.Group("/parameters")
		{
			parameters.POST("", r.createParameter)
			parameters.GET("", r.getAllParameters)
			parameters.GET("/:id", r.getParameterByID)
			parameters.PATCH("/:id", r.updateParameter)
			parameters.PUT("/:id", r.updateParameterWithRules)
			parameters.DELETE("/:id", r.deleteParameter)
			parameters.POST("/simulate", r.simulateParameter)
		}

		// Experiment routes
		experiments := v1.Group("/experiments")
		{
			experiments.POST("", r.createExperiment)
			experiments.GET("", r.getAllExperiments)
			experiments.GET("/:id", r.getExperimentByID)
			experiments.PATCH("/:id/reject", r.rejectExperiment)
			experiments.PATCH("/:id/approve", r.approveExperiment)
			experiments.PATCH("/:id/abort", r.abortExperiment)
		}

		// SDK routes
		sdk := v1.Group("/sdk")
		{
			sdk.POST("/metadata", r.getMetadataSDK)
			sdk.POST("/parameters", r.getAllParametersSDK)
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
