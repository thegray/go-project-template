package logger

import (
	"context"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const TraceIDHeader = "X-Trace-Id"

type traceIDKey struct{}

type Logger struct {
	*zap.Logger
}

func New(env string) (*zap.Logger, error) {
	fmt.Println("logger", env)
	if env == "production" {
		return zap.NewProduction()
	}
	return zap.NewDevelopment()
}

func Wrap(log *zap.Logger) *Logger {
	if log == nil {
		log = zap.NewNop()
	}
	return &Logger{Logger: log}
}

func Layer(log *zap.Logger, name string) *Logger {
	if log == nil {
		log = zap.NewNop()
	}
	if name == "" {
		return Wrap(log)
	}
	return Wrap(log.Named(name))
}

func (l *Logger) Named(name string) *Logger {
	if l == nil || l.Logger == nil {
		return Wrap(nil)
	}
	return Wrap(l.Logger.Named(name))
}

func (l *Logger) WithCtx(ctx context.Context) *Logger {
	return Wrap(WithContext(l.base(), ctx))
}

func (l *Logger) DebugCtx(ctx context.Context, msg string, fields ...zap.Field) {
	l.base().With(FieldsFromContext(ctx)...).Debug(msg, fields...)
}

func (l *Logger) InfoCtx(ctx context.Context, msg string, fields ...zap.Field) {
	l.base().With(FieldsFromContext(ctx)...).Info(msg, fields...)
}

func (l *Logger) WarnCtx(ctx context.Context, msg string, fields ...zap.Field) {
	l.base().With(FieldsFromContext(ctx)...).Warn(msg, fields...)
}

func (l *Logger) ErrorCtx(ctx context.Context, msg string, fields ...zap.Field) {
	l.base().With(FieldsFromContext(ctx)...).Error(msg, fields...)
}

func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, traceIDKey{}, traceID)
}

func TraceID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	if value, ok := ctx.Value(traceIDKey{}).(string); ok {
		return value
	}
	return ""
}

func FieldsFromContext(ctx context.Context) []zap.Field {
	traceID := TraceID(ctx)
	if traceID == "" {
		return nil
	}

	return []zap.Field{zap.String("trace_id", traceID)}
}

func WithContext(log *zap.Logger, ctx context.Context) *zap.Logger {
	if log == nil {
		log = zap.NewNop()
	}

	fields := FieldsFromContext(ctx)
	if len(fields) == 0 {
		return log
	}

	return log.With(fields...)
}

func GinMiddleware(log *Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		base := log.base()
		traceID := c.GetHeader(TraceIDHeader)
		if traceID == "" {
			traceID = uuid.NewString()
		}

		c.Writer.Header().Set(TraceIDHeader, traceID)
		c.Request = c.Request.WithContext(WithTraceID(c.Request.Context(), traceID))

		start := time.Now()
		c.Next()

		fields := []zap.Field{
			zap.String("trace_id", traceID),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("latency", time.Since(start)),
			zap.String("ip", c.ClientIP()),
		}

		if len(c.Errors) > 0 {
			base.Error("request completed with errors", append(fields, zap.String("errors", c.Errors.String()))...)
			return
		}

		base.Info("request completed", fields...)
	}
}

func (l *Logger) base() *zap.Logger {
	if l == nil || l.Logger == nil {
		return zap.NewNop()
	}
	return l.Logger
}
