/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements. See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership. The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied. See the License for the
 *  specific language governing permissions and limitations
 *  under the License.
 */

package org.apache.custos.scim.resource.manager;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Component;
import org.wso2.charon3.core.exceptions.*;
import org.wso2.charon3.core.extensions.UserManager;
import org.wso2.charon3.core.objects.Group;
import org.wso2.charon3.core.objects.User;
import org.wso2.charon3.core.utils.CopyUtil;
import org.wso2.charon3.core.utils.codeutils.SearchRequest;

import java.util.List;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;

/**
 * Class responsible for manage Users. Responsible for request response formatting and
 * interact with core services
 */
@Component
public class ResourceManager implements UserManager {

    private static final Logger logger = LoggerFactory.getLogger(ResourceManager.class);
    //in memory user manager stores users
    ConcurrentHashMap<String, User> inMemoryUserList = new ConcurrentHashMap<String, User>();
    ConcurrentHashMap<String, Group> inMemoryGroupList = new ConcurrentHashMap<String, Group>();
    @Override
    public User createUser(User user, Map<String, Boolean> map) throws CharonException, ConflictException, BadRequestException {
        if (inMemoryUserList.get(user.getId()) != null) {
            throw new ConflictException("User with the id : " + user.getId() + "already exists");
        } else {
            inMemoryUserList.put(user.getId(), user);
            return (User) CopyUtil.deepCopy(user);
        }
    }

    @Override
    public User getUser(String id, Map<String, Boolean> map) throws CharonException, BadRequestException, NotFoundException {
        if (inMemoryUserList.get(id) != null) {
            return (User) CopyUtil.deepCopy(inMemoryUserList.get(id));
        } else {
            throw new NotFoundException("No user with the id : " + id);
        }
    }

    @Override
    public void deleteUser(String id) throws NotFoundException, CharonException, NotImplementedException, BadRequestException {
        if (inMemoryUserList.get(id) == null) {
            throw new NotFoundException("No user with the id : " + id);
        } else {
            inMemoryUserList.remove(id);
        }
    }

    @Override
    public List<Object> listUsersWithPost(SearchRequest searchRequest, Map<String, Boolean> map) throws CharonException, NotImplementedException, BadRequestException {
        return null;
    }

    @Override
    public User updateUser(User user, Map<String, Boolean> map) throws NotImplementedException, CharonException, BadRequestException, NotFoundException {
        if (user.getId() != null) {
            inMemoryUserList.replace(user.getId(), user);
            return (User) CopyUtil.deepCopy(user);
        } else {
            throw new NotFoundException("No user with the id : " + user.getId());
        }
    }

    @Override
    public User getMe(String s, Map<String, Boolean> map) throws CharonException, BadRequestException, NotFoundException {
        return null;
    }

    @Override
    public User createMe(User user, Map<String, Boolean> map) throws CharonException, ConflictException, BadRequestException {
        return null;
    }

    @Override
    public void deleteMe(String s) throws NotFoundException, CharonException, NotImplementedException, BadRequestException {

    }

    @Override
    public User updateMe(User user, Map<String, Boolean> map) throws NotImplementedException, CharonException, BadRequestException, NotFoundException {
        return null;
    }

    @Override
    public Group createGroup(Group group, Map<String, Boolean> map) throws CharonException, ConflictException, NotImplementedException, BadRequestException {
        return null;
    }

    @Override
    public Group getGroup(String s, Map<String, Boolean> map) throws NotImplementedException, BadRequestException, CharonException, NotFoundException {
        return null;
    }

    @Override
    public void deleteGroup(String s) throws NotFoundException, CharonException, NotImplementedException, BadRequestException {

    }

    @Override
    public Group updateGroup(Group group, Group group1, Map<String, Boolean> map) throws NotImplementedException, BadRequestException, CharonException, NotFoundException {
        return null;
    }

    @Override
    public List<Object> listGroupsWithPost(SearchRequest searchRequest, Map<String, Boolean> map) throws NotImplementedException, BadRequestException, CharonException {
        return null;
    }
}
