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
package org.apache.custos.amie.service;

import org.apache.custos.amie.model.ClusterAccountEntity;
import org.apache.custos.amie.model.ProjectEntity;
import org.apache.custos.amie.model.ProjectMembershipEntity;
import org.apache.custos.amie.repo.ClusterAccountRepository;
import org.apache.custos.amie.repo.ProjectMembershipRepository;
import org.apache.custos.amie.repo.ProjectRepository;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.List;
import java.util.Optional;

@Service
public class ProjectMembershipService {

    private static final Logger LOGGER = LoggerFactory.getLogger(ProjectMembershipService.class);

    private final ProjectMembershipRepository membershipRepository;
    private final ProjectRepository projectRepository;
    private final ClusterAccountRepository clusterAccountRepository;

    public ProjectMembershipService(ProjectMembershipRepository membershipRepository, ProjectRepository projectRepository, ClusterAccountRepository clusterAccountRepository) {
        this.membershipRepository = membershipRepository;
        this.projectRepository = projectRepository;
        this.clusterAccountRepository = clusterAccountRepository;
    }

    /**
     * Creates a membership linking a cluster account to a project with a role.
     *
     * @param projectId        The project ID.
     * @param clusterAccountId The cluster account ID.
     * @param role             The role (PI, USER).
     * @return The created membership entity.
     */
    @Transactional
    public ProjectMembershipEntity createMembership(String projectId, String clusterAccountId, String role) {
        Optional<ProjectMembershipEntity> existing = membershipRepository.findByProjectIdAndClusterAccountId(projectId, clusterAccountId);
        if (existing.isPresent()) {
            LOGGER.info("Membership already exists for project [{}] and cluster account [{}]", projectId, clusterAccountId);
            return existing.get();
        }

        ProjectEntity project = projectRepository.findById(projectId)
                .orElseThrow(() -> new IllegalArgumentException("Project not found: " + projectId));

        ClusterAccountEntity clusterAccount = clusterAccountRepository.findById(clusterAccountId)
                .orElseThrow(() -> new IllegalArgumentException("Cluster account not found: " + clusterAccountId));

        ProjectMembershipEntity membership = new ProjectMembershipEntity();
        membership.setProject(project);
        membership.setClusterAccount(clusterAccount);
        membership.setRole(role);
        membership.setActive(true);

        membershipRepository.save(membership);
        LOGGER.info("Created membership for project [{}], cluster account [{}] with role [{}]", projectId, clusterAccountId, role);

        return membership;
    }

    /**
     * Inactivates a specific membership (user removed from project).
     *
     * @param projectId The project ID.
     * @param personId  Person ID
     */
    @Transactional
    public void inactivateMembershipsByPersonAndProject(String projectId, String personId) {
        // TODO - If the user is a PI of a project?
        //  - right now only the membership is turned inactive, no changes to the project
        List<ProjectMembershipEntity> memberships = membershipRepository.findByProjectIdAndClusterAccount_Person_Id(projectId, personId);

        if (memberships.isEmpty()) {
            LOGGER.warn("No memberships found for person [{}] on project [{}]. No action taken.", personId, projectId);
            return;
        }

        memberships.forEach(membership -> membership.setActive(false));
        membershipRepository.saveAll(memberships);
        LOGGER.info("Inactivated {} membership(s) for person [{}] on project [{}]", memberships.size(), personId, projectId);
    }

    /**
     * Inactivates all memberships for a project.
     *
     * @param projectId The project ID.
     */
    @Transactional
    public void inactivateAllMembershipsForProject(String projectId) {
        List<ProjectMembershipEntity> projectMemberships = membershipRepository.findByProjectId(projectId);

        if (projectMemberships.isEmpty()) {
            LOGGER.info("No memberships to inactivate for project [{}]", projectId);
            return;
        }

        projectMemberships.forEach(membership -> membership.setActive(false));
        membershipRepository.saveAll(projectMemberships);

        LOGGER.info("Inactivated {} memberships for project [{}]", projectMemberships.size(), projectId);
    }

    /**
     * Reactivates PI membership for a project.
     *
     * @param projectId The project ID.
     */
    @Transactional
    public void reactivatePiMembership(String projectId) {
        List<ProjectMembershipEntity> piMemberships = membershipRepository.findByProjectIdAndRole(projectId, "PI");
        piMemberships.forEach(membership -> membership.setActive(true));
        membershipRepository.saveAll(piMemberships);
        LOGGER.info("Reactivated {} PI memberships for project [{}]", piMemberships.size(), projectId);
    }

}
