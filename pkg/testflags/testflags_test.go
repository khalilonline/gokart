package testflags

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEvaluate(t *testing.T) {
	t.Run("should skip", func(t *testing.T) {
		t.Setenv("TEST_TYPE", string(Unit))

		done := make(chan struct{})
		tt := &testing.T{}
		go func(t *testing.T) {
			defer close(done)
			Evaluate(t, Integration)
		}(tt)

		<-done
		assert.True(t, tt.Skipped())
	})

	t.Run("should run", func(t *testing.T) {
		t.Setenv("TEST_TYPE", string(Integration))

		done := make(chan struct{})
		var executed bool

		tt := &testing.T{}
		go func(t *testing.T) {
			defer close(done)
			Evaluate(t, Integration)
			executed = true
		}(tt)

		<-done
		assert.True(t, executed)
	})

	t.Run("with user defined test type", func(t *testing.T) {
		var readerIntegration TestType = "reader_integration"

		t.Run("should skip", func(t *testing.T) {
			t.Setenv("TEST_TYPE", string(Unit))

			done := make(chan struct{})
			tt := &testing.T{}
			go func(t *testing.T) {
				defer close(done)
				Evaluate(t, readerIntegration, Integration)
			}(tt)

			<-done
			assert.True(t, tt.Skipped())
		})

		t.Run("should run when 'system'", func(t *testing.T) {
			t.Setenv("TEST_TYPE", string(Integration))

			done := make(chan struct{})
			var executed bool

			tt := &testing.T{}
			go func(t *testing.T) {
				defer close(done)
				Evaluate(t, readerIntegration, Integration)
				executed = true
			}(tt)

			<-done
			assert.True(t, executed)
		})

		t.Run("should run when 'reader_integration'", func(t *testing.T) {
			t.Setenv("TEST_TYPE", string(readerIntegration))

			done := make(chan struct{})
			var executed bool

			tt := &testing.T{}
			go func(t *testing.T) {
				defer close(done)
				Evaluate(t, readerIntegration, Integration)
				executed = true
			}(tt)

			<-done
			assert.True(t, executed)
		})
	})
}
