/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements. See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership. The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License. You may obtain a copy of the License at
 *
 *  http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing,
 *  software distributed under the License is distributed on an
 *  "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 *  KIND, either express or implied. See the License for the
 *  specific language governing permissions and limitations
 *  under the License.
 */

import {
    useToast,
    Button, Input, Modal,
    ModalOverlay,
    ModalContent,
    ModalHeader,
    ModalFooter,
    ModalBody,
    ModalCloseButton, FormControl,
    FormLabel,
    Stack,
    Select
} from "@chakra-ui/react";
import { BACKEND_URL } from "../../lib/constants";
import axios, { AxiosResponse } from "axios";
import React from "react";
import { useAuth } from "react-oidc-context";
import { useNavigate } from "react-router-dom";
  
  export const AddGroupMemberModal = ({
    groupId,
    isOpen,
    onClose,
  }: {
    groupId: string | undefined;
    isOpen: boolean;
    onClose: () => void;
  }) => {
    const toast = useToast();
    const [newEmail, setNewEmail] = React.useState("");
    const [newRole, setNewRole] = React.useState("MEMBER");
    const auth = useAuth();
    const navigate = useNavigate();

    const handleAddMember = async () => {
      const emailRegex = /^[a-zA-Z0-9._-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,6}$/;
  
      if (!emailRegex.test(newEmail)) {
        toast({
          title: "Invalid email",
          status: "error",
          duration: 2000,
          isClosable: true,
        });
        return;
      }
  
      const resp = await axios
        .post(
          `${BACKEND_URL}/api/v1/group-management/groups/${groupId}/members`,
          {
            username: newEmail,
            type: newRole,
          },
          {
            headers: {
              Authorization: `Bearer ${auth.user?.access_token}`,
            },
          }
        )
        .catch((error) => {
          toast({
            title: "Error adding member",
            description: error.response.data.error,
            status: "error",
            duration: 2000,
            isClosable: true,
          });
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        }) as AxiosResponse<any, any>;
      
      if (resp.status > 199 && resp.status < 300) {
        onClose();
        navigate(0);
      }
    };
  
    return (
      <>
        <Modal isOpen={isOpen} onClose={onClose}>
          <ModalOverlay />
          <ModalContent>
            <ModalHeader>Add Member</ModalHeader>
            <ModalCloseButton />
            <ModalBody>
              <Stack spacing={4}>
                <FormControl>
                  <FormLabel>Email</FormLabel>
                  <Input
                    type="email"
                    placeholder="name@example.com"
                    value={newEmail}
                    onChange={(e) => setNewEmail(e.target.value)}
                  />
                </FormControl>
  
                <FormControl>
                  <FormLabel>Role</FormLabel>
                  <Select
                    value={newRole}
                    onChange={(e) => setNewRole(e.target.value)}
                  >
                    <option value="MEMBER">Member</option>
                    <option value="OWNER">Owner</option>
                    <option value="ADMIN">ADMIN</option>
                  </Select>
                </FormControl>
  
                <Button colorScheme="blue" onClick={handleAddMember}>
                  Add
                  {newEmail && newRole ? ` ${newEmail} as ${newRole}` : ""}
                </Button>
              </Stack>
            </ModalBody>
  
            <ModalFooter />
          </ModalContent>
        </Modal>
      </>
    );
  };
  