package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	Log *zap.Logger
)

// common field keys
const (
	UserIDKey    = "user_id"
	DeviceToken  = "device_toke"
	EmailKey     = "email"
	SessionIDKey = "session_id"
	IPKey        = "ip"
	ErrorKey     = "error"
	CallerKey    = "caller"
)

func init() {
	var err error
	config := zap.NewProductionConfig()

	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.StacktraceKey = ""
	encoderConfig.CallerKey = CallerKey
	encoderConfig.MessageKey = "message"
	encoderConfig.LevelKey = "level"

	config.EncoderConfig = encoderConfig
	config.OutputPaths = []string{"stdout"}
	config.ErrorOutputPaths = []string{"stderr"}

	Log, err = config.Build(zap.AddCallerSkip(1))
	if err != nil {
		panic(err)
	}
}

// fields type for more convenient logging
type Fields map[string]interface{}

// info level with logs optional fields
func Info(msg string, fields ...Fields) {
	if len(fields) > 0 {
		Log.Info(msg, getZapFields(fields[0])...)
		return
	}
	Log.Info(msg)
}

// error level logs with optional fields
func Error(msg string, fields ...Fields) {
	if len(fields) > 0 {
		Log.Error(msg, getZapFields(fields[0])...)
		return
	}
	Log.Error(msg)
}

// debug level logs with optional fields
func Debug(msg string, fields ...Fields) {
	if len(fields) > 0 {
		Log.Debug(msg, getZapFields(fields[0])...)
		return
	}
	Log.Debug(msg)
}

// warn level logs with optional fields
func Warn(msg string, fields ...Fields) {
	if len(fields) > 0 {
		Log.Warn(msg, getZapFields(fields[0])...)
		return
	}
	Log.Warn(msg)
}

// fatal level logs with optional fields, calls os.Exit(1)
func Fatal(msg string, fields ...Fields) {
	if len(fields) > 0 {
		Log.Fatal(msg, getZapFields(fields[0])...)
		return
	}

	Log.Fatal(msg)
}

// adds an error field to the log entry
func WithError(err error) Fields {
	return Fields{
		ErrorKey: err.Error(),
	}
}

// adds a user ID field to the log entry
func WithUserID(userID string) Fields {
	return Fields{
		UserIDKey: userID,
	}
}

// adds an email field to the log entry
func WithEmail(email string) Fields {
	return Fields{
		EmailKey: email,
	}
}

// adds a session ID field to the log entry
func WithSessionID(sessionID string) Fields {
	return Fields{
		SessionIDKey: sessionID,
	}
}

// adds an IP address field to the log entry
func WithIP(ip string) Fields {
	return Fields{
		IPKey: ip,
	}
}

// adds an device token field to the log entry
func WithDeviceToken(token string) Fields {
	return Fields{
		DeviceToken: token,
	}
}

// combines multiple fields into a single Fields object
func Merge(fields ...Fields) Fields {
	merged := make(Fields)
	for _, f := range fields {
		for k, v := range f {
			merged[k] = v
		}
	}
	return merged
}

// convert fields to zap.Field slice
func getZapFields(fields Fields) []zap.Field {
	zapFields := make([]zap.Field, 0, len(fields))
	for k, v := range fields {
		zapFields = append(zapFields, zap.Any(k, v))
	}
	return zapFields
}
