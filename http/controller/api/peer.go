package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	requstform "github.com/lejianwen/rustdesk-api/v2/http/request/api"
	"github.com/lejianwen/rustdesk-api/v2/http/response"
	"github.com/lejianwen/rustdesk-api/v2/service"
	"net/http"
)

type Peer struct {
}

// SysInfo
// @Tags System
// @Summary 提交系统信息
// @Description 提交系统信息
// @Accept  json
// @Produce  json
// @Param body body requstform.PeerForm true "系统信息表单"
// @Success 200 {string} string "SYSINFO_UPDATED,ID_NOT_FOUND"
// @Failure 500 {object} response.ErrorResponse
// @Router /sysinfo [post]
func (p *Peer) SysInfo(c *gin.Context) {
	f := &requstform.PeerForm{}
	err := c.ShouldBindBodyWith(f, binding.JSON)
	if err != nil {
		response.Error(c, response.TranslateMsg(c, "ParamsError")+err.Error())
		return
	}
	fpe := f.ToPeer()
	pe := service.AllService.PeerService.FindById(f.Id)
	if pe.RowId == 0 {
		// id未找到：可能是设备改了id或MachineGuid(uuid)后重新上报。
		// 按CPU、主机名、内存三个维度匹配旧记录，命中则视为同一台设备，
		// 覆盖旧记录(含id、uuid)，并同步修正地址簿中的旧id，避免出现重复设备。
		old := service.AllService.PeerService.FindByHardware(f.Cpu, f.Hostname, f.Memory)
		if old.RowId > 0 {
			fpe.RowId = old.RowId
			fpe.UserId = old.UserId
			if fpe.UserId == 0 {
				fpe.UserId = service.AllService.UserService.FindLatestUserIdFromLoginLogByUuid(fpe.Uuid, fpe.Id)
			}
			err = service.AllService.PeerService.TakeOver(old, fpe)
			if err != nil {
				response.Error(c, response.TranslateMsg(c, "OperationFailed")+err.Error())
				return
			}
			c.String(http.StatusOK, "SYSINFO_UPDATED")
			return
		}
		pe = f.ToPeer()
		pe.UserId = service.AllService.UserService.FindLatestUserIdFromLoginLogByUuid(pe.Uuid, pe.Id)
		err = service.AllService.PeerService.Create(pe)
		if err != nil {
			response.Error(c, response.TranslateMsg(c, "OperationFailed")+err.Error())
			return
		}
	} else {
		if pe.UserId == 0 {
			pe.UserId = service.AllService.UserService.FindLatestUserIdFromLoginLogByUuid(pe.Uuid, pe.Id)
		}
		fpe.RowId = pe.RowId
		fpe.UserId = pe.UserId
		err = service.AllService.PeerService.Update(fpe)
		if err != nil {
			response.Error(c, response.TranslateMsg(c, "OperationFailed")+err.Error())
			return
		}
	}
	//SYSINFO_UPDATED 上传成功
	//ID_NOT_FOUND 下次心跳会上传
	//直接响应文本
	c.String(http.StatusOK, "SYSINFO_UPDATED")
}

// SysInfoVer
// @Tags System
// @Summary 获取系统版本信息
// @Description 获取系统版本信息
// @Accept  json
// @Produce  json
// @Success 200 {string} string ""
// @Failure 500 {object} response.ErrorResponse
// @Router /sysinfo_ver [post]
func (p *Peer) SysInfoVer(c *gin.Context) {
	//读取resources/version文件
	v := service.AllService.AppService.GetAppVersion()
	// 加上启动时间，方便client上传信息
	v = fmt.Sprintf("%s\n%s", v, service.AllService.AppService.GetStartTime())
	c.String(http.StatusOK, v)
}
