package zaplogger

import (
	"go.uber.org/zap"
	"testing"
	"time"
)

func TestNewZapLogger(t *testing.T) {
	runConf := NewRunConf("debug", "dev")
	logger := NewZapLogger(ConfigZap("test", runConf))

	t.Run("simple_test", func(t *testing.T) {
		logger.Info("test Info",
			zap.String("a", "aaa"),
			zap.Int("b", 3),
			zap.Duration("c", time.Second),
		)
	})

}
