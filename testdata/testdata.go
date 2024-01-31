// Copyright 2023 Kapeta Inc.
// SPDX-License-Identifier: MIT

package testdata

import _ "embed"

//go:embed cluster-service.yml
var ClusterServiceYml []byte

//go:embed block.yml
var BlockYml []byte
