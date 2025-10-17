/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements. See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership. The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License. You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */
package org.apache.custos.access.ci.service.service;

import org.apache.custos.access.ci.service.model.ProjectEntity;
import org.apache.custos.access.ci.service.repo.ProjectRepository;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.Optional;

@Service
public class ProjectService {

    private static final Logger LOGGER = LoggerFactory.getLogger(ProjectService.class);

    private final ProjectRepository projectRepository;

    public ProjectService(ProjectRepository projectRepository) {
        this.projectRepository = projectRepository;
    }

    /**
     * Creates a new project or returns an existing one if found.
     *
     * @param projectId   The project ID from AMIE.
     * @param grantNumber The grant number from AMIE.
     * @return The project entity (created or existing).
     */
    @Transactional
    public ProjectEntity createOrFindProject(String projectId, String grantNumber) {
        Optional<ProjectEntity> existing = projectRepository.findById(projectId);
        if (existing.isPresent()) {
            LOGGER.info("Project [{}] already exists with grant number [{}]", projectId, existing.get().getGrantNumber());
            return existing.get();
        }

        ProjectEntity project = new ProjectEntity();
        project.setId(projectId);
        project.setGrantNumber(grantNumber);
        project.setActive(true);
        projectRepository.save(project);

        LOGGER.info("Created new project [{}] with grant number [{}]", projectId, grantNumber);
        return project;
    }

    /**
     * Inactivates a project.
     *
     * @param projectId The project ID to inactivate.
     */
    @Transactional
    public void inactivateProject(String projectId) {
        ProjectEntity project = projectRepository.findById(projectId).orElseThrow(() -> new IllegalArgumentException("Project not found: " + projectId));

        project.setActive(false);
        projectRepository.save(project);
        LOGGER.info("Inactivated project [{}]", projectId);
    }

    /**
     * Reactivates a project.
     *
     * @param projectId The project ID to reactivate.
     */
    @Transactional
    public void reactivateProject(String projectId) {
        ProjectEntity project = projectRepository.findById(projectId).orElseThrow(() -> new IllegalArgumentException("Project not found: " + projectId));

        project.setActive(true);
        projectRepository.save(project);
        LOGGER.info("Reactivated project [{}]", projectId);
    }

}
