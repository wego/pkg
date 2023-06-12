package wegin

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

// New returns a new gin engine with custom validators， used to replace gin.New()
func New() *gin.Engine {
	e := gin.New()
	registerValidator()
	return e
}

// Default returns a new gin engine with custom validators， used to replace gin.Default()
func Default() *gin.Engine {
	e := gin.Default()
	registerValidator()
	return e
}

func registerValidator() {
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
}
