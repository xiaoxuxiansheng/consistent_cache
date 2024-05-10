package runtime

import (
	"fmt"
	"os"
	"runtime"
	"strings"
)

// 生成由当前进程 id 和协程 id 组成的标识字符串
func GetCurrentProcessAndGogroutineIDStr() string {
	pid := GetCurrentProcessID()
	goroutineID := GetCurrentGoroutineID()
	return fmt.Sprintf("%d_%s", pid, goroutineID)
}

// GetCurrentGoroutineID 获取当前的协程ID
func GetCurrentGoroutineID() string {
	buf := make([]byte, 128)
	buf = buf[:runtime.Stack(buf, false)]
	stackInfo := string(buf)
	return strings.TrimSpace(strings.Split(strings.Split(stackInfo, "[running]")[0], "goroutine")[1])
}

// 获取当前的进程ID
func GetCurrentProcessID() int {
	return os.Getpid()
}
