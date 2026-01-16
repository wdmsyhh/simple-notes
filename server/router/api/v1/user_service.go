package v1

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	apiv1 "github.com/wdmsyhh/simple-notes/proto/gen/api/v1"
	storepb "github.com/wdmsyhh/simple-notes/proto/gen/store"
	"github.com/wdmsyhh/simple-notes/server/auth"
	"github.com/wdmsyhh/simple-notes/service"
	"github.com/wdmsyhh/simple-notes/store"
)

// RegisterUser 注册新用户
func (s *APIV1Service) RegisterUser(ctx context.Context, request *apiv1.RegisterUserRequest) (*storepb.User, error) {
	// 验证请求
	if request.User == nil {
		return nil, status.Errorf(codes.InvalidArgument, "user is required")
	}
	if request.Password == "" {
		return nil, status.Errorf(codes.InvalidArgument, "password is required")
	}

	// 检查用户注册是否启用
	// 目前我们假设它总是启用的
	// 稍后我们将添加设置来控制这一点

	// 通过服务层创建用户
	regReq := &service.UserRegistrationRequest{
		Username: request.User.Username,
		Password: request.Password,
		Nickname: request.User.Nickname,
		Avatar:   request.User.Avatar,
		Bio:      request.User.Bio,
	}

	user, err := s.userService.RegisterUser(ctx, regReq)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to register user: %v", err)
	}

	// 转换为 protobuf 消息
	return convertUserToProto(user), nil
}

// LoginUser 用户登录
func (s *APIV1Service) LoginUser(ctx context.Context, request *apiv1.LoginUserRequest) (*apiv1.LoginUserResponse, error) {
	// 验证请求
	if request.Username == "" || request.Password == "" {
		return nil, status.Errorf(codes.InvalidArgument, "username and password are required")
	}

	// 认证用户
	loginReq := &service.UserLoginRequest{
		Username: request.Username,
		Password: request.Password,
	}

	user, err := s.userService.LoginUser(ctx, loginReq)
	if err != nil {
		// 返回友好的中文错误信息
		return nil, status.Errorf(codes.Unauthenticated, "用户名或密码错误")
	}

	// 生成认证令牌
	roleStr := convertUserRoleToProto(user.Role).String()
	accessToken, _, err := auth.GenerateAccessTokenV2(int32(user.ID), user.Username, roleStr, []byte(s.Secret))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate access token: %v", err)
	}

	return &apiv1.LoginUserResponse{
		User:  convertUserToProto(user),
		Token: accessToken,
	}, nil
}

// GetUser 根据ID获取用户
func (s *APIV1Service) GetUser(ctx context.Context, request *apiv1.GetUserRequest) (*storepb.User, error) {
	// 从资源名称中提取用户ID
	userID, err := extractUserIDFromName(request.Name)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user name: %v", err)
	}

	// 从服务层获取用户
	user, err := s.userService.GetUserByID(ctx, userID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get user: %v", err)
	}
	if user == nil {
		return nil, status.Errorf(codes.NotFound, "user not found")
	}

	return convertUserToProto(user), nil
}

// GetCurrentUser 获取当前已认证的用户
func (s *APIV1Service) GetCurrentUser(ctx context.Context, request *apiv1.GetCurrentUserRequest) (*storepb.User, error) {
	// 从上下文中获取用户声明
	claims := auth.GetUserClaims(ctx)
	if claims == nil {
		return nil, status.Errorf(codes.Unauthenticated, "authentication required")
	}

	// 从存储层获取用户
	user, err := s.userService.GetUserByID(ctx, uint(claims.UserID))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get user: %v", err)
	}
	if user == nil {
		return nil, status.Errorf(codes.NotFound, "user not found")
	}

	return convertUserToProto(user), nil
}

// UpdateUser 更新用户
func (s *APIV1Service) UpdateUser(ctx context.Context, request *apiv1.UpdateUserRequest) (*storepb.User, error) {
	// 从资源名称中提取用户ID
	userID, err := extractUserIDFromName(request.User.Name)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user name: %v", err)
	}

	// 获取当前用户
	currentUser, err := s.userService.GetUserByID(ctx, userID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get user: %v", err)
	}
	if currentUser == nil {
		return nil, status.Errorf(codes.NotFound, "user not found")
	}

	// 检查权限
	// 目前，只有用户自己可以更新他们的个人资料
	// 稍后，我们将为管理员实现适当的权限检查
	// currentAuthUser, _ := s.fetchCurrentUser(ctx)
	// if currentAuthUser == nil || currentAuthUser.ID != userID {
	//     return nil, status.Errorf(codes.PermissionDenied, "permission denied")
	// }

	// 更新用户字段
	if request.User.Username != "" {
		currentUser.Username = request.User.Username
	}
	if request.User.Nickname != "" {
		currentUser.Nickname = request.User.Nickname
	}
	if request.User.Avatar != "" {
		currentUser.Avatar = request.User.Avatar
	}
	if request.User.Bio != "" {
		currentUser.Bio = request.User.Bio
	}
	if request.User.Role != storepb.UserRole_USER_ROLE_UNSPECIFIED {
		currentUser.Role = convertUserRoleFromProto(request.User.Role)
	}

	// 通过服务层更新用户
	updatedUser, err := s.userService.UpdateUser(ctx, currentUser)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update user: %v", err)
	}

	return convertUserToProto(updatedUser), nil
}

// DeleteUser 删除用户
func (s *APIV1Service) DeleteUser(ctx context.Context, request *apiv1.DeleteUserRequest) (*emptypb.Empty, error) {
	// 从资源名称中提取用户ID
	userID, err := extractUserIDFromName(request.Name)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user name: %v", err)
	}

	// 检查权限
	// 目前，只有用户自己可以删除他们的账户
	// 稍后，我们将为管理员实现适当的权限检查
	// currentAuthUser, _ := s.fetchCurrentUser(ctx)
	// if currentAuthUser == nil || currentAuthUser.ID != userID {
	//     return nil, status.Errorf(codes.PermissionDenied, "permission denied")
	// }

	// 通过服务层删除用户
	err = s.userService.DeleteUser(ctx, userID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete user: %v", err)
	}

	return &emptypb.Empty{}, nil
}

// ListUsers 列出用户
func (s *APIV1Service) ListUsers(ctx context.Context, request *apiv1.ListUsersRequest) (*apiv1.ListUsersResponse, error) {
	// 检查权限
	// 只有管理员可以列出用户
	// currentAuthUser, _ := s.fetchCurrentUser(ctx)
	// if currentAuthUser == nil || !service.IsSuperUser(currentAuthUser) {
	//     return nil, status.Errorf(codes.PermissionDenied, "permission denied")
	// }

	// 通过服务层获取用户
	users, err := s.userService.ListUsers(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list users: %v", err)
	}

	// 转换为 protobuf 消息
	protoUsers := make([]*storepb.User, len(users))
	for i, user := range users {
		protoUsers[i] = convertUserToProto(user)
	}

	return &apiv1.ListUsersResponse{
		Users:    protoUsers,
		Total:    int32(len(users)),
		Page:     request.Page,
		PageSize: request.PageSize,
	}, nil
}

// 辅助函数

// extractUserIDFromName 从资源名称中提取用户ID
func extractUserIDFromName(name string) (uint, error) {
	parts := strings.Split(name, "/")
	if len(parts) != 2 || parts[0] != "users" {
		return 0, fmt.Errorf("invalid resource name: %s", name)
	}

	id, err := strconv.ParseUint(parts[1], 10, 32)
	if err != nil {
		return 0, fmt.Errorf("invalid user ID: %s", parts[1])
	}

	return uint(id), nil
}

// convertUserToProto 将 store.User 转换为 storepb.User
func convertUserToProto(user *store.User) *storepb.User {
	return &storepb.User{
		Name:         fmt.Sprintf("users/%d", user.ID),
		Id:           int64(user.ID),
		Username:     user.Username,
		PasswordHash: user.PasswordHash,
		Nickname:     user.Nickname,
		Avatar:       user.Avatar,
		Bio:          user.Bio,
		Role:         convertUserRoleToProto(user.Role),
		CreatedAt:    user.CreatedAt.Unix(),
		UpdatedAt:    user.UpdatedAt.Unix(),
	}
}

// convertUserRoleToProto 将 store.UserRole 转换为 storepb.UserRole
func convertUserRoleToProto(role store.UserRole) storepb.UserRole {
	switch role {
	case store.RoleHost:
		return storepb.UserRole_USER_ROLE_HOST
	case store.RoleAdmin:
		return storepb.UserRole_USER_ROLE_ADMIN
	case store.RoleUser:
		return storepb.UserRole_USER_ROLE_USER
	default:
		return storepb.UserRole_USER_ROLE_UNSPECIFIED
	}
}

// convertUserRoleFromProto 将 storepb.UserRole 转换为 store.UserRole
func convertUserRoleFromProto(role storepb.UserRole) store.UserRole {
	switch role {
	case storepb.UserRole_USER_ROLE_HOST:
		return store.RoleHost
	case storepb.UserRole_USER_ROLE_ADMIN:
		return store.RoleAdmin
	case storepb.UserRole_USER_ROLE_USER:
		return store.RoleUser
	default:
		return store.RoleUser
	}
}

// fetchCurrentUser 从上下文中检索当前已认证的用户
func (s *APIV1Service) fetchCurrentUser(ctx context.Context) (*store.User, error) {
	claims := auth.GetUserClaims(ctx)
	if claims == nil {
		return nil, nil
	}

	user, err := s.userService.GetUserByID(ctx, uint(claims.UserID))
	if err != nil {
		return nil, err
	}

	return user, nil
}
