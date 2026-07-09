// Copyright 2026 sigma
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package mcpserver

import (
	"context"

	"github.com/go-sigma/sigma/pkg/dal/models"
)

type contextKey string

const (
	contextKeyUser      contextKey = "sigma.mcp.user"
	contextKeyClientIP  contextKey = "sigma.mcp.client_ip"
	contextKeyUserAgent contextKey = "sigma.mcp.user_agent"
)

func withUser(ctx context.Context, user *models.User) context.Context {
	return context.WithValue(ctx, contextKeyUser, user)
}

func userFromContext(ctx context.Context) (*models.User, bool) {
	user, ok := ctx.Value(contextKeyUser).(*models.User)
	return user, ok && user != nil
}

func withClientInfo(ctx context.Context, clientIP, userAgent string) context.Context {
	ctx = context.WithValue(ctx, contextKeyClientIP, clientIP)
	return context.WithValue(ctx, contextKeyUserAgent, userAgent)
}

func clientInfoFromContext(ctx context.Context) (string, string) {
	clientIP, _ := ctx.Value(contextKeyClientIP).(string)
	userAgent, _ := ctx.Value(contextKeyUserAgent).(string)
	return clientIP, userAgent
}
