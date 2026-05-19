// Licensed to the Apache Software Foundation (ASF) under one
// or more contributor license agreements.  See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership.  The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License.  You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

// This file is the build-time selector for which connectors are bundled into
// the binary. Each blank import below registers a connector via its init()
// function. To exclude a connector from a deployment, comment out its line.
// To add a new connector, drop in a new blank import.

package main

import (
	_ "github.com/apache/airavata-custos/connectors/ACCESS/AMIE-Processor"
)
