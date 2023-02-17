package ctxflow

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/go-ins/ctxflow/layer"
	"github.com/go-ins/ctxflow/puzzle"
	"runtime"

	"github.com/gin-gonic/gin"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"reflect"
)

func slave(src interface{}) interface{} {
	typ := reflect.TypeOf(src)
	if typ.Kind() == reflect.Ptr { //如果是指针类型
		typ = typ.Elem()                              //获取源实际类型(否则为指针类型)
		dst := reflect.New(typ).Elem()                //创建对象
		b, _ := json.Marshal(src)                     //导出json
		_ = json.Unmarshal(b, dst.Addr().Interface()) //json序列化
		return dst.Addr().Interface()                 //返回指针
	} else {
		dst := reflect.New(typ).Elem()                //创建对象
		b, _ := json.Marshal(src)                     //导出json
		_ = json.Unmarshal(b, dst.Addr().Interface()) //json序列化
		return dst.Interface()                        //返回值
	}
}

func UseController(controller layer.IController) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		ctl := slave(controller).(layer.IController)
		ctl.SetContext(ctx)

		logCtx := puzzle.LogCtx{
			LogId:   puzzle.GetLogID(ctx),
			ReqId:   puzzle.GetRequestID(ctx),
			AppName: puzzle.GetAppName(),
			LocalIp: puzzle.GetLocalIp(),
		}
		ctl.SetLogCtx(&logCtx)
		ctl.SetLog(puzzle.GetDefaultSugaredLogger().With(
			zap.String("logId", logCtx.LogId),
			zap.String("requestId", logCtx.ReqId),
			zap.String("module", logCtx.AppName),
			zap.String("localIp", logCtx.LocalIp),
		))

		defer NoPanicContorller(ctl)

		ctl.PreUse()
		ctl.Action()
	}
}

func NoPanicContorller(ctl layer.IController) {
	if err := recover(); err != nil {
		stack := PanicTrace(4) //4KB
		ctl.GetLog().Errorf("[controller panic]:%+v,stack:%s", err, string(stack))
		ctl.RenderJsonFail(errors.New("异常错误！"))
	}
}

func NoPanic(flow layer.IFlow) {
	if err := recover(); err != nil {
		stack := PanicTrace(4) //4KB
		flow.GetLog().Errorf("[controller panic]:%+v,stack:%s", err, string(stack))
	}
}

func StackTrace() []byte {
	e := []byte("\ngoroutine ")
	line := []byte("\n")
	stack := make([]byte, 4<<10) //4KB
	length := runtime.Stack(stack, true)
	stack = stack[0:length]
	start := bytes.Index(stack, line) + 1
	stack = stack[start:]
	end := bytes.LastIndex(stack, line)
	if end != -1 {
		stack = stack[:end]
	}
	end = bytes.Index(stack, e)
	if end != -1 {
		stack = stack[:end]
	}
	stack = bytes.TrimRight(stack, "\n")

	return stack
}

func PanicTrace(kb int) []byte {
	s := []byte("/src/runtime/panic.go")
	e := []byte("\ngoroutine ")
	line := []byte("\n")
	stack := make([]byte, kb<<10) //4KB
	length := runtime.Stack(stack, true)
	start := bytes.Index(stack, s)
	stack = stack[start:length]
	start = bytes.Index(stack, line) + 1
	stack = stack[start:]
	end := bytes.LastIndex(stack, line)
	if end != -1 {
		stack = stack[:end]
	}
	end = bytes.Index(stack, e)
	if end != -1 {
		stack = stack[:end]
	}
	stack = bytes.TrimRight(stack, "\n")
	return stack
}

func UseTask(task layer.ITask) func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		ctx := &gin.Context{}
		task2 := slave(task).(layer.ITask)
		task2.SetContext(ctx)
		logCtx := puzzle.LogCtx{
			LogId:   puzzle.GetLogID(ctx),
			ReqId:   puzzle.GetRequestID(ctx),
			AppName: puzzle.GetAppName(),
			LocalIp: puzzle.GetLocalIp(),
		}
		task2.SetLogCtx(&logCtx)
		task2.SetLog(puzzle.GetDefaultSugaredLogger().With(
			zap.String("logId", logCtx.LogId),
			zap.String("requestId", logCtx.ReqId),
			zap.String("module", logCtx.AppName),
			zap.String("localIp", logCtx.LocalIp),
		))
		task2.PreUse()
		task2.Run(args)
	}
}

func UseCron(cron layer.ICrontab) func() {
	return func() {
		ctx := &gin.Context{}
		cr := slave(cron).(layer.ICrontab)
		cr.SetContext(ctx)
		logCtx := puzzle.LogCtx{
			LogId:   puzzle.GetLogID(ctx),
			ReqId:   puzzle.GetRequestID(ctx),
			AppName: puzzle.GetAppName(),
			LocalIp: puzzle.GetLocalIp(),
		}
		cr.SetLogCtx(&logCtx)
		cr.SetLog(puzzle.GetDefaultSugaredLogger().With(
			zap.String("logId", logCtx.LogId),
			zap.String("requestId", logCtx.ReqId),
			zap.String("module", logCtx.AppName),
			zap.String("localIp", logCtx.LocalIp),
		))
		cr.PreUse()
		cr.Run()
	}
}
