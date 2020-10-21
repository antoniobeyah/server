// Copyright (c) 2020 Target Brands, Inc. All rights reserved.
//
// Use of this source code is governed by the LICENSE file in this repository.

package router

import (
	"github.com/gin-gonic/gin"
	"github.com/go-vela/server/api"
	"github.com/go-vela/server/router/middleware"
	"github.com/go-vela/server/router/middleware/repo"
)

// CompileHandlers is a function that extends the provided base router group
// with the API handlers for compile functionality.
//
// POST   /api/v1/compile
func CompileHandlers(base *gin.RouterGroup) {
	// Compile endpoints
	compile := base.Group("/compile/:org/:repo", repo.Establish())
	{
		compile.POST("", middleware.Payload(), api.Compile)
	} // end of compile endpoints
}
