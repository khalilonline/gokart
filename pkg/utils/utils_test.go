package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type UtilsTestSuite struct {
	suite.Suite
}

func TestUtilsTestSuite(t *testing.T) {
	suite.Run(t, new(UtilsTestSuite))
}

func (suite *UtilsTestSuite) TestSafeDeref() {
	suite.T().Run("string pointer", func(t *testing.T) {
		str := "Hello"
		result := SafeDeref(&str)
		if result != "Hello" {
			suite.T().Errorf("Expected 'Hello', got '%v'", result)
		}
	})

	suite.T().Run("nil string pointer", func(t *testing.T) {
		var nilStr *string
		result := SafeDeref(nilStr)
		if result != "" {
			suite.T().Errorf("Expected '', got '%v'", result)
		}
	})

	suite.T().Run("int pointer", func(t *testing.T) {
		num := 1
		result := SafeDeref(&num)
		if result != 1 {
			suite.T().Errorf("Expected 1, got %v", result)
		}
	})

	suite.T().Run("nil int pointer", func(t *testing.T) {
		var nilInt *int
		result := SafeDeref(nilInt)
		if result != 0 {
			suite.T().Errorf("Expected 0, got %v", result)
		}
	})

	suite.T().Run("bool pointer", func(t *testing.T) {
		b := true
		result := SafeDeref(&b)
		if result != true {
			suite.T().Errorf("Expected true, got %v", result)
		}
	})

	suite.T().Run("nil bool pointer", func(t *testing.T) {
		var nilBool *bool
		result := SafeDeref(nilBool)
		if result != false {
			suite.T().Errorf("Expected false, got %v", result)
		}
	})

	suite.T().Run("float64 pointer", func(t *testing.T) {
		f := 3.14
		result := SafeDeref(&f)
		if result != 3.14 {
			suite.T().Errorf("Expected 3.14, got %v", result)
		}
	})

	suite.T().Run("nil float64 pointer", func(t *testing.T) {
		var nilFloat *float64
		result := SafeDeref(nilFloat)
		if result != 0.0 {
			suite.T().Errorf("Expected 0.0, got %v", result)
		}
	})
}

func (suite *UtilsTestSuite) TestAsPtrFromAny() {
	var nilValue any
	result := AsPtrFromAny[string](nilValue)
	if result != nil {
		suite.T().Errorf("Expected nil, but got: %v", result)
	}

	strValue := "hello"
	result = AsPtrFromAny[string](strValue)
	if result == nil || *result != strValue {
		suite.T().Errorf("Expected pointer to 'hello', but got: %v", result)
	}

	intValue := 1
	intResult := AsPtrFromAny[int](intValue)
	if intResult == nil || *intResult != intValue {
		suite.T().Errorf("Expected pointer to 1, but got: %v", intResult)
	}
}

func (suite *UtilsTestSuite) TestAsValueFromAny() {
	var nilValue any
	result, ok := AsValueFromAny[string](nilValue)
	if ok {
		suite.T().Errorf("Expected ok to be false for nil value, but got true")
	}
	if result != "" {
		suite.T().Errorf("Expected empty string for nil value, but got: %v", result)
	}

	strValue := "hello"
	strResult, ok := AsValueFromAny[string](strValue)
	if !ok {
		suite.T().Errorf("Expected ok to be true for string value, but got false")
	}
	if strResult != strValue {
		suite.T().Errorf("Expected 'hello', but got: %v", strResult)
	}

	intValue := 1
	intResult, ok := AsValueFromAny[int](intValue)
	if !ok {
		suite.T().Errorf("Expected ok to be true for int value, but got false")
	}
	if intResult != intValue {
		suite.T().Errorf("Expected 1, but got: %v", intResult)
	}
}

func TestToStrings(t *testing.T) {
	t.Run("slice of underlying string type", func(t *testing.T) {
		type SomeStringType string

		assert.Equal(t, []string{"a", "b", "c"}, ToStrings([]SomeStringType{"a", "b", "c"}))
	})

	t.Run("slice of strings", func(t *testing.T) {
		assert.Equal(t, []string{"a", "b", "c"}, ToStrings([]string{"a", "b", "c"}))
	})
}
