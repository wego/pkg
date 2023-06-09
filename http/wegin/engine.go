package wegin

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

// New returns a new gin engine with custom validators
func New() *gin.Engine {
	e := gin.New()
	for key, value := range fieldValidators {
		if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
			_ = v.RegisterValidation(key, value)
		}
	}
	for key, value := range structValidators {
		if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
			v.RegisterStructValidation(value, key)
		}
	}
	return e
}
