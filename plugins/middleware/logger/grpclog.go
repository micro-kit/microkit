package logger

import (
	"go.uber.org/zap"
)

/* 设置grpc内置日志对象 */

// GrpcLog 符合grpclog.Logger接口
type GrpcLog struct {
	SugaredLogger *zap.SugaredLogger
}

// Info logs to INFO log. Arguments are handled in the manner of fmt.Print.
func (l *GrpcLog) Info(args ...interface{}) {
	// l.SugaredLogger.Info(args...)
}

// Infoln logs to INFO log. Arguments are handled in the manner of fmt.Println.
func (l *GrpcLog) Infoln(args ...interface{}) {
	// args = append(args, "\n")
	// l.SugaredLogger.Info(args...)
}

// Infof logs to INFO log. Arguments are handled in the manner of fmt.Printf.
func (l *GrpcLog) Infof(format string, args ...interface{}) {
	// l.SugaredLogger.Infof(format, args...)
}

// Warning logs to WARNING log. Arguments are handled in the manner of fmt.Print.
func (l *GrpcLog) Warning(args ...interface{}) {
	l.SugaredLogger.Warn(args...)
}

// Warningln logs to WARNING log. Arguments are handled in the manner of fmt.Println.
func (l *GrpcLog) Warningln(args ...interface{}) {
	args = append(args, "\n")
	l.SugaredLogger.Warn(args...)
}

// Warningf logs to WARNING log. Arguments are handled in the manner of fmt.Printf.
func (l *GrpcLog) Warningf(format string, args ...interface{}) {
	l.SugaredLogger.Warnf(format, args...)
}

// Error logs to ERROR log. Arguments are handled in the manner of fmt.Print.
func (l *GrpcLog) Error(args ...interface{}) {
	l.SugaredLogger.Error(args...)
}

// Errorln logs to ERROR log. Arguments are handled in the manner of fmt.Println.
func (l *GrpcLog) Errorln(args ...interface{}) {
	args = append(args, "\n")
	l.SugaredLogger.Error(args...)
}

// Errorf logs to ERROR log. Arguments are handled in the manner of fmt.Printf.
func (l *GrpcLog) Errorf(format string, args ...interface{}) {
	l.SugaredLogger.Errorf(format, args...)
}

// Fatal logs to ERROR log. Arguments are handled in the manner of fmt.Print.
// gRPC ensures that all Fatal logs will exit with os.Exit(1).
// Implementations may also call os.Exit() with a non-zero exit code.
func (l *GrpcLog) Fatal(args ...interface{}) {
	l.SugaredLogger.Fatal(args...)
}

// Fatalln logs to ERROR log. Arguments are handled in the manner of fmt.Println.
// gRPC ensures that all Fatal logs will exit with os.Exit(1).
// Implementations may also call os.Exit() with a non-zero exit code.
func (l *GrpcLog) Fatalln(args ...interface{}) {
	args = append(args, "\n")
	l.SugaredLogger.Fatal(args...)
}

// Fatalf logs to ERROR log. Arguments are handled in the manner of fmt.Printf.
// gRPC ensures that all Fatal logs will exit with os.Exit(1).
// Implementations may also call os.Exit() with a non-zero exit code.
func (l *GrpcLog) Fatalf(format string, args ...interface{}) {
	l.SugaredLogger.Fatalf(format, args...)
}

// V reports whether verbosity level l is at least the requested verbose level.
func (*GrpcLog) V(l int) bool {
	return true
}
