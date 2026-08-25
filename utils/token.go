package utils

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
)

type TokenClaims struct {
	UserId int64 `json:"user_id"`
	jwt.RegisteredClaims
}

// 生成token
func GenerateToken(ctx context.Context, userId int64) (string, error) {
	if userId <= 0 {
		return "", fmt.Errorf("用户ID无效")
	}
	if Red == nil {
		return "", fmt.Errorf("Redis未初始化")
	}

	// jwt加密的密钥
	secret := viper.GetString("jwt.secret")
	// jwt的签发者
	issuer := viper.GetString("jwt.issuer")
	// jwt的过期时间
	expire := viper.GetDuration("jwt.expire")

	// 生成长度为16的byte数组 - 使用rand.Read生成随机数
	randomBytes := make([]byte, 16)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", fmt.Errorf("生成会话ID失败： %w", err)
	}

	// 将随机数数组转成会话的id - 十六进制
	tokenId := hex.EncodeToString(randomBytes)

	now := time.Now()           // 获取当前时间
	expireAt := now.Add(expire) //获取过期时间

	claims := TokenClaims{
		UserId: userId,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    issuer,
			Subject:   strconv.FormatInt(userId, 10),
			ID:        tokenId,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expireAt),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(secret))

	if err != nil {
		return "", fmt.Errorf("JWT签名失败：%w", err)
	}

	sessionKey := fmt.Sprintf("im:auth:session:%s", tokenId)
	userSessionKey := fmt.Sprintf("im:auth:user:%d:session", userId)

	pipe := Red.TxPipeline()
	// 保存当个session
	pipe.Set(ctx, sessionKey, strconv.FormatInt(userId, 10), time.Until(expireAt))
	// 将session_id 存入到用户的会话列表中去 - 多设备情况
	pipe.SAdd(ctx, userSessionKey, tokenId)
	pipe.ExpireAt(ctx, userSessionKey, expireAt)

	if _, err := pipe.Exec(ctx); err != nil {
		return "", fmt.Errorf("保存登录会话失败：%w", err)
	}

	return signedToken, nil
}

// 校验token
func ValidateToken(ctx context.Context, rawToken string) (*TokenClaims, error) {
	rawToken = strings.TrimSpace(rawToken)
	if rawToken == "" {
		return nil, fmt.Errorf("Token不能为空")
	}
	if Red == nil {
		return nil, fmt.Errorf("Redis未初始化")
	}

	secret := viper.GetString("jwt.secret")
	issuer := viper.GetString("jwt.issuer")

	claims := &TokenClaims{}

	token, err := jwt.ParseWithClaims(rawToken, claims,
		func(t *jwt.Token) (interface{}, error) {
			if t.Method != jwt.SigningMethodHS256 {
				return nil, fmt.Errorf("参数有误")
			}
			return []byte(secret), nil
		}, jwt.WithValidMethods([]string{
			jwt.SigningMethodHS256.Alg(),
		}), jwt.WithIssuer(issuer), jwt.WithExpirationRequired(), jwt.WithIssuedAt(), jwt.WithLeeway(30*time.Second))
	if err != nil {
		return nil, fmt.Errorf("Token校验失败")
	}

	if !token.Valid {
		return nil, fmt.Errorf("Token无效")
	}

	if claims.UserId <= 0 {
		return nil, fmt.Errorf("Token中的用户id无效")
	}

	if claims.ID == "" {
		return nil, fmt.Errorf("Token缺少会话ID")
	}
	if claims.Subject == "" {
		return nil, fmt.Errorf("Token缺少用户标识")
	}

	subjectUserId, err := strconv.ParseInt(claims.Subject, 10, 64)
	if err != nil || subjectUserId != claims.UserId {
		return nil, fmt.Errorf("Token校驗失敗")
	}

	sessionKey := fmt.Sprintf("im:auth:session:%s", claims.ID)

	sessionUserId, err := Red.Get(ctx, sessionKey).Result()
	if errors.Is(err, redis.Nil) {
		return nil, fmt.Errorf("登錄會話已過期或已退出")
	}
	if err != nil {
		return nil, fmt.Errorf("查詢登錄會話失敗：%w", err)
	}
	if sessionUserId != strconv.FormatInt(claims.UserId, 10) {
		return nil, fmt.Errorf("Token校验失败")
	}
	return claims, nil
}

// 撤销一个登录会话
func RevokeToken(ctx context.Context, userId int64, tokenId string) error {
	if userId <= 0 || tokenId == "" {
		return fmt.Errorf("参数有误")
	}
	sessionKey := fmt.Sprintf("im:auth:session:%s", tokenId)
	userSessionKey := fmt.Sprintf("im:auth:user:%d:session", userId)

	pipe := Red.TxPipeline()
	// 删除指定的token
	pipe.Del(ctx, sessionKey)
	// 从session_id 列表中移除要删除的
	pipe.SRem(ctx, userSessionKey, tokenId)

	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("撤销登录会话失败：%w", err)
	}

	return nil
}

// 撤销某个用户所有登录会话
func RevokeUserTokens(ctx context.Context, userId int64) error {
	if userId <= 0 {
		return fmt.Errorf("参数有误")
	}

	userSessionKey := fmt.Sprintf("im:auth:user:%d:session", userId)
	// 获取列表中的数据
	tokenIds, err := Red.SMembers(ctx, userSessionKey).Result()
	if err != nil {
		return fmt.Errorf("查询用户会话失败：%w", err)
	}

	if len(tokenIds) == 0 {
		return nil
	}

	keys := make([]string, 0, len(tokenIds)+1)

	for _, tokenId := range tokenIds {
		keys = append(keys, fmt.Sprintf("im:auth:session:%s", tokenId))
	}
	keys = append(keys, userSessionKey)

	if err := Red.Del(ctx, keys...).Err(); err != nil {
		return fmt.Errorf("撤销用户全部会话失败，%w", err)
	}

	return nil
}
