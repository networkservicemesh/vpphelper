// Copyright (c) 2024 Cisco and/or its affiliates.
//
// SPDX-License-Identifier: Apache-2.0
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package extendtimeout - provides a wrapper for vpp connection that uses extended timeout for all vpp operations
package extendtimeout

import (
	"context"
	"time"

	"github.com/networkservicemesh/sdk/pkg/tools/extend"
	"github.com/networkservicemesh/sdk/pkg/tools/log"
	"go.fd.io/govpp/api"
)

type extendedConnection struct {
	api.Connection
	contextTimeout time.Duration
}

// NewConnection - creates a wrapper for vpp connection that uses extended context timeout for all operations
func NewConnection(vppConn api.Connection, contextTimeout time.Duration) api.Connection {
	return &extendedConnection{
		Connection:     vppConn,
		contextTimeout: contextTimeout,
	}
}

func (c *extendedConnection) Invoke(ctx context.Context, req, reply api.Message) error {
	ctx, cancel := c.withExtendedTimeoutCtx(ctx)
	err := c.Connection.Invoke(ctx, req, reply)
	cancel()
	return err
}

func (c *extendedConnection) withExtendedTimeoutCtx(ctx context.Context) (extendedCtx context.Context, cancel func()) {
	deadline, ok := ctx.Deadline()
	if !ok {
		return ctx, func() {}
	}

	minDeadline := time.Now().Add(c.contextTimeout)
	if minDeadline.Before(deadline) {
		return ctx, func() {}
	}
	log.FromContext(ctx).Warn("Context deadline has been extended by extendtimeout from %v to %v", deadline, minDeadline)
	deadline = minDeadline
	postponedCtx, cancel := context.WithDeadline(context.Background(), deadline)
	return extend.WithValuesFromContext(postponedCtx, ctx), cancel
}
