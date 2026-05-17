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

// Package service provides high-level operations on the core domain entities
// (Organization, User, Project). It hides database transactions and other
// persistence concerns from callers, so other modules can use simple
// CreateX / GetX / UpdateX / DeleteX functions.
package service

import "errors"

// ErrNotFound is returned when a requested record does not exist.
var ErrNotFound = errors.New("record not found")

// ErrAlreadyExists is returned when attempting to create a record that
// conflicts with an existing one (e.g. duplicate email).
var ErrAlreadyExists = errors.New("record already exists")

// ErrInvalidInput is returned when required fields are missing or invalid.
var ErrInvalidInput = errors.New("invalid input")
