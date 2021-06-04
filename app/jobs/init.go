/*
 * @Author: hc
 * @Date: 2021-06-01 15:28:28
 * @LastEditors: hc
 * @LastEditTime: 2021-06-03 10:32:31
 * @Description:
 */
package jobs

import (
	"example-webcron/app/models"
	"fmt"
	"os/exec"
	"time"

	"github.com/astaxie/beego"
)

func InitJobs() {
	condition := make(map[string]interface{})
	condition["status"] = 1
	list, _ := models.TaskGetList(1, 1000000, condition)
	for _, task := range list {
		job, err := NewJobFromTask(task)
		if err != nil {
			beego.Error("InitJobs:", err.Error())
			continue
		}
		AddJob(task.CronSpec, job)
	}
}

func runCmdWithTimeout(cmd *exec.Cmd, timeout time.Duration) (error, bool) {
	done := make(chan error)
	go func() {
		done <- cmd.Wait()
	}()

	var err error
	select {
	case <-time.After(timeout):
		beego.Warn(fmt.Sprintf("任务执行时间超过%d秒，进程将被强制杀掉: %d", int(timeout/time.Second), cmd.Process.Pid))
		go func() {
			<-done // 读出上面的goroutine数据，避免阻塞导致无法退出
		}()
		if err = cmd.Process.Kill(); err != nil {
			beego.Error(fmt.Sprintf("进程无法杀掉: %d, 错误信息: %s", cmd.Process.Pid, err))
		}
		return err, true
	case err = <-done:
		return err, false
	}
}
