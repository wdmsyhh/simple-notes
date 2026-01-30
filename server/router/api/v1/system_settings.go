package v1

import (
	"encoding/json"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/wdmsyhh/simple-notes/server/auth"
	"github.com/wdmsyhh/simple-notes/store"
)

// getSystemSettingsHandler 返回系统设置 JSON（如 login_enabled），供前端判断是否显示登录入口
func (s *APIV1Service) getSystemSettingsHandler(c echo.Context) error {
	ctx := c.Request().Context()
	loginEnabled, err := s.Store.GetLoginEnabled(ctx)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to get settings"})
	}
	return c.JSON(http.StatusOK, map[string]bool{"login_enabled": loginEnabled})
}

// setSystemSettingsRequest 设置系统设置请求体
type setSystemSettingsRequest struct {
	LoginEnabled *bool `json:"login_enabled"`
}

// setSystemSettingsHandler 设置系统设置，仅 HOST 角色可调用
func (s *APIV1Service) setSystemSettingsHandler(c echo.Context) error {
	ctx := c.Request().Context()
	authHeader := c.Request().Header.Get("Authorization")
	authenticator := auth.NewAuthenticator(s.Store, s.Secret)
	result := authenticator.Authenticate(ctx, authHeader)
	if result == nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "未登录或 token 无效"})
	}
	user, err := s.userService.GetUserByID(ctx, uint(result.Claims.UserID))
	if err != nil || user == nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "用户不存在"})
	}
	if user.Role != store.RoleHost {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "仅系统管理员可修改此设置"})
	}
	var body setSystemSettingsRequest
	if err := json.NewDecoder(c.Request().Body).Decode(&body); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "无效的请求体"})
	}
	if body.LoginEnabled != nil {
		if err := s.Store.SetLoginEnabled(ctx, *body.LoginEnabled); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "保存设置失败"})
		}
	}
	loginEnabled, _ := s.Store.GetLoginEnabled(ctx)
	return c.JSON(http.StatusOK, map[string]bool{"login_enabled": loginEnabled})
}
