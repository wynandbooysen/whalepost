package main

// whalepost
// Copyright (C) 2018 Maximilian Pachl

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

// ---------------------------------------------------------------------------------------
//  imports
// ---------------------------------------------------------------------------------------

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/faryon93/util"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// ---------------------------------------------------------------------------------------
//  types
// ---------------------------------------------------------------------------------------

// UpdateBody is the users request to update a service image.
type UpdateBody struct {
	Image string `json:"image" schema:"image"`
}

// UpdateResponse is returned to the user upon success.
type UpdateResponse struct {
	Status string `json:"status"`
	Image  string `json:"image"`
}

// ---------------------------------------------------------------------------------------
//  imports
// ---------------------------------------------------------------------------------------

// ServiceUpdate handels the update request of a swarm service.
func ServiceUpdate(w http.ResponseWriter, r *http.Request) {
	serviceId := mux.Vars(r)["ServiceId"]

	start := time.Now()

	// parse the request body
	var body UpdateBody
	err := util.ParseBody(r, &body)
	if err != nil {
		http.Error(w, "body: "+err.Error(), http.StatusBadRequest)
		return
	}

	// TODO: choose api version automatically
	docker, err := client.NewClientWithOpts(client.WithHost(Endpoint), client.WithVersion(ApiVersion))
	if err != nil {
		logrus.Errorln("failed to create docker client:", err.Error())
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	// fetch the current service sepcs
	ctx := context.Background()
	opt := types.ServiceInspectOptions{}
	service, _, err := docker.ServiceInspectWithRaw(ctx, serviceId, opt)
	if client.IsErrNotFound(err) {
		logrus.Errorln("failed to inspect service:", err.Error())
		http.Error(w, "no such service", http.StatusNotFound)
		return
	} else if err != nil {
		logrus.Errorln("failed to inspect service:", err.Error())
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	// make sure that service updates are allowed
	allow := strings.ToLower(service.Spec.Labels[LabelAllow])
	if allow != "true" && allow != "yes" && allow != "on" {
		logrus.Errorln("rejecting update: service is not allowed to be updated")
		http.Error(w, "service update to allowed", http.StatusForbidden)
		return
	}

	// if a new image has been requests -> insert it into the new container spec
	if body.Image != "" {
		logrus.Infof("replacing image \"%s\" with \"%s\"",
			service.Spec.TaskTemplate.ContainerSpec.Image, body.Image)
		service.Spec.TaskTemplate.ContainerSpec.Image = body.Image
	}

	// update the service
	updateOpts := types.ServiceUpdateOptions{
		QueryRegistry: true,
	}
	resp, err := docker.ServiceUpdate(ctx, serviceId, service.Version, service.Spec, updateOpts)
	if err != nil {
		logrus.Errorln("failed to update service:", err.Error())
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	// display the warnings returend by docker
	for _, warn := range resp.Warnings {
		logrus.Warnln("dockerd:", warn)
	}

	// tell the user that everything is fine
	logrus.Infof("updated service \"%s\" to image \"%s\" (%s)",
		serviceId, service.Spec.TaskTemplate.ContainerSpec.Image, time.Since(start))

	util.Jsonify(w, UpdateResponse{
		Status: "success",
		Image:  service.Spec.TaskTemplate.ContainerSpec.Image,
	})
}
