package redisclient

import (
	"context"

	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/mock"
)

type MockRedisClient struct {
	mock.Mock
}

func (m *MockRedisClient) XAdd(ctx context.Context, a *redis.XAddArgs) *redis.StringCmd {
	args := m.Called(ctx, a)
	return args.Get(0).(*redis.StringCmd)
}

func (m *MockRedisClient) XGroupCreateMkStream(ctx context.Context, stream, group, start string) *redis.StatusCmd {
	args := m.Called(ctx, stream, group, start)
	return args.Get(0).(*redis.StatusCmd)
}

func (m *MockRedisClient) XReadGroup(ctx context.Context, a *redis.XReadGroupArgs) *redis.XStreamSliceCmd {
	args := m.Called(ctx, a)
	return args.Get(0).(*redis.XStreamSliceCmd)
}

func (m *MockRedisClient) XAck(ctx context.Context, stream string, group string, ids ...string) *redis.IntCmd {
	args := m.Called(ctx, stream, group, ids)
	return args.Get(0).(*redis.IntCmd)
}
