// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/backend/blob/main/LICENSE)

package endpoints

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/fasthttp/router"
	"github.com/tharsis/dashboard-backend/internal/db"
	"github.com/tharsis/dashboard-backend/internal/requester"
	"github.com/tharsis/dashboard-backend/internal/resources"
	"github.com/valyala/fasthttp"
)

func NetworkConfig(ctx *fasthttp.RequestCtx) {
	networkConfigs, err := resources.GetNetworkConfigs()
	if err != nil {
		sendResponse(buildErrorResponse(err.Error()), nil, ctx)
		return
	}
	stringRes, err := json.Marshal(networkConfigs)
	if err != nil {
		sendResponse("", err, ctx)
		return
	}

	res := "{\"values\":" + string(stringRes) + "}"

	sendResponse(res, nil, ctx)
}

type SourceParams struct {
	SourceChannel         string `json:"sourceChannel"`
	DestinationChannel    string `json:"destinationChannel"`
	SourceIBCDenomToEvmos string `json:"sourceIBCDenomToEvmos"`
}

type ChainConfigParams struct {
	ConfigurationType string       `json:"configurationType"`
	Source            SourceParams `json:"source"`
	ClientID          string       `json:"clientId"`
	ChainID           string       `json:"chainId"`
	ExplorerTxURL     string       `json:"explorerTxUrl"`
}

type ConfigurationParams struct {
	Configurations []ChainConfigParams `json:"configurations"`
	Prefix         string              `json:"prefix"`
}

type NetworkByName struct {
	Values ConfigurationParams `json:"values"`
}

func NetworkConfigByNameInternal(name string) (string, error) {
	name = strings.ToLower(name)

	if val, err := db.RedisGetNetworkConfigByName(name); err == nil {
		return val, nil
	}

	val, err := requester.GetNetworkConfig()
	if err != nil {
		return "", err
	}

	for _, v := range val {
		if strings.Contains(v.URL, name) {
			res := "{\"values\":" + v.Content + "}"
			db.RedisSetNetworkConfigByName(name, res)
			return res, nil
		}
	}
	return "", fmt.Errorf("invalid network")
}

func NetworkConfigByName(ctx *fasthttp.RequestCtx) {
	name := paramToString("name", ctx)

	val, err := NetworkConfigByNameInternal(name)
	if err != nil {
		sendResponse(buildErrorResponse(err.Error()), nil, ctx)
		return
	}
	sendResponse(val, nil, ctx)
}

func AddNetworkRoutes(r *router.Router) {
	r.GET("/NetworkConfig", NetworkConfig)
	r.GET("/NetworkConfig/{name}", NetworkConfigByName)
}
