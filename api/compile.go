// Copyright (c) 2020 Target Brands, Inc. All rights reserved.
//
// Use of this source code is governed by the LICENSE file in this repository.

package api

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/go-vela/compiler/compiler"
	"github.com/go-vela/server/database"
	"github.com/go-vela/server/router/middleware/repo"
	"github.com/go-vela/server/source"
	"github.com/go-vela/server/util"
	"github.com/go-vela/types"
	"github.com/go-vela/types/yaml"
	"github.com/sirupsen/logrus"
)

// swagger:operation GET /compile router Compile
//
// Retrieve a compiled Vela pipeline
//
// ---
// x-success_http_code: '200'
// produces:
// - application/json
// parameters:
// responses:
//   '200':
//     description: Successfully 'ping'-ed Vela API
//     schema:
//       type: string

// Compile represents the API handler to
// retrieve a compiled Vela pipeline.
func Compile(c *gin.Context) {

	r := repo.Retrieve(c)

	// send API call to capture the repo owner
	u, err := database.FromContext(c).GetUser(r.GetUserID())
	if err != nil {
		retErr := fmt.Errorf("unable to get owner for %s: %w", r.GetFullName(), err)

		util.HandleError(c, http.StatusBadRequest, retErr)

		return
	}

	// templatesCommit := "d9dd52256c576b3b5809c65af3fe99862cedd805"
	// simpleCommit := "5b9fe91dd5ffa704b156e141e9382fb06d5a3a04"

	cm := os.Getenv("VLCM")

	// send API call to capture the pipeline configuration file
	config, err := source.FromContext(c).ConfigBackoff(u, r.GetOrg(), r.GetName(), cm)

	// send API call to capture the pipeline configuration file
	if err != nil {
		retErr := fmt.Errorf("unable to get pipeline configuration: %w", err)

		util.HandleError(c, http.StatusNotFound, retErr)

		return
	}

	// capture middleware values
	m := c.MustGet("metadata").(*types.Metadata)

	// parse and compile the pipeline configuration file
	p, err := compiler.FromContext(c).
		WithMetadata(m).
		WithUser(u).Parse(config)
	if err != nil {
		retErr := fmt.Errorf("unable to parse pipeline configuration: %w", err)

		util.HandleError(c, http.StatusInternalServerError, retErr)

		return
	}

	// validate the yaml configuration
	err = compiler.FromContext(c).
		WithMetadata(m).
		WithUser(u).Validate(p)
	if err != nil {

		retErr := fmt.Errorf("unable to validate pipeline: %w", err)

		util.HandleError(c, http.StatusInternalServerError, retErr)
		return
	}

	// create map of templates for easy lookup
	tmpls := mapFromTemplates(p.Templates)
	if len(p.Stages) > 0 {

		// inject the templates into the stages
		p.Stages, err = compiler.FromContext(c).
			WithMetadata(m).
			WithUser(u).ExpandStages(p.Stages, tmpls)
		if err != nil {
			retErr := fmt.Errorf("unable to expand stages: %w", err)

			util.HandleError(c, http.StatusInternalServerError, retErr)

			return
		}
	}

	logrus.Info("tmpls")
	logrus.Info(tmpls)

	// _, err = c.PrivateGithub.Template(c.user, src)
	// inject the templates into the steps
	p.Steps, err = compiler.FromContext(c).
		WithMetadata(m).
		WithUser(u).ExpandSteps(p.Steps, tmpls)

	if err != nil {
		retErr := fmt.Errorf("unable to expand steps: %w", err)

		util.HandleError(c, http.StatusInternalServerError, retErr)

		return
	}
	c.JSON(http.StatusOK, p)

}

// helper function that creates a map of templates from a yaml configuration.
func mapFromTemplates(templates []*yaml.Template) map[string]*yaml.Template {
	m := make(map[string]*yaml.Template)

	for _, tmpl := range templates {
		m[tmpl.Name] = tmpl
	}

	return m
}
