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
  Box,
  Flex,
  FormControl,
  Text,
  FormLabel,
  Input,
  Stack,
  Button,
  Table,
  Thead,
  Tbody,
  Tr,
  Th,
  Td,
  TableContainer,
  Code,
  IconButton,
  Switch,
  useToast,
  useDisclosure,
} from "@chakra-ui/react";
import { PageTitle } from "../PageTitle";
import { FiTrash2 } from "react-icons/fi";
import { ActionButton } from "../ActionButton";
import { BACKEND_URL } from "../../lib/constants";
import { useEffect } from "react";
import React from "react";
import { useAuth } from "react-oidc-context";
import { Group, Member } from "../../interfaces/Groups";
import { useNavigate } from "react-router-dom";
import axios from "axios";
import { TransferOwnershipModal } from "./TransferOwnershipModal";
import { LeftRightLayout } from "../LeftRightLayout";
import { StackedBorderBox } from "../StackedBorderBox";

interface GroupSettingsProps {
  groupId: string | undefined;
}

export const GroupSettings = ({ groupId }: GroupSettingsProps) => {
  const [name, setName] = React.useState("");
  const [description, setDescription] = React.useState("");
  const [owner, setOwner] = React.useState("");
  const [groupManagers, setGroupManagers] = React.useState([]);
  const [roles, setRoles] = React.useState([] as string[] | undefined);
  const [hasOwnerPriv, setHasOwnerPriv] = React.useState(false);
  const auth = useAuth();
  const toast = useToast();
  const navigate = useNavigate();
  const transferOwnershipDisclosure = useDisclosure();

  const customFetch = async (url: string, options?: RequestInit) => {
    const resp = await fetch(url, {
      ...options,
      headers: {
        ...options?.headers,
        Authorization: `Bearer ${auth?.user?.access_token}`,
      },
    });

    const data = await resp.json();

    return data;
  };

  useEffect(() => {
    (async () => {
      const groupBasicInfo: Group = await customFetch(
        `${BACKEND_URL}/api/v1/group-management/groups/${groupId}`
      );
      setName(groupBasicInfo.name);
      setDescription(groupBasicInfo.description);
      setOwner(groupBasicInfo.owner_id);
      setRoles(groupBasicInfo.client_roles);

      const groupMembers = await customFetch(
        `${BACKEND_URL}/api/v1/group-management/groups/${groupId}/members`
      );
      const groupManagers = groupMembers.profiles.filter(
        (member: Member) => member.membership_type === "ADMIN"
      );
      setGroupManagers(groupManagers);

      let userRole = "";
      let hasOwnerPower = false;

      userRole = groupMembers.profiles.filter(
        (member: Member) => member.email === auth.user?.profile.email
      )[0]?.membership_type;

      if (userRole) {
        hasOwnerPower = userRole.toUpperCase() === "OWNER";
      }

      setHasOwnerPriv(hasOwnerPower);
    })();
  }, []);

  const handleSaveChanges = async () => {
    console.log(name, description);

    const resp = await customFetch(
      `${BACKEND_URL}/api/v1/group-management/groups/${groupId}`,
      {
        method: "PUT",
        body: JSON.stringify({
          name,
          description,
        }),
        headers: {
          "Content-Type": "application/json",
        },
      }
    );

    console.log(resp);

    if (resp?.id) {
      // was successful
      navigate(0);
    } else {
      toast({
        title: "Could not save group",
        status: "error",
        duration: 3000,
        isClosable: true,
      });
    }
  };

  if (!groupId) {
    return;
  }

  return (
    <>
      <PageTitle size="md">Group Settings</PageTitle>
      <Text color="default.secondary">
        Edit group membership, roles, and other information.
      </Text>

      <StackedBorderBox>
        <LeftRightLayout
          left={<Text fontSize="lg">Basic Information</Text>}
          right={
            <Stack spacing={4}>
              <FormControl color="default.default">
                <FormLabel>Name</FormLabel>
                <Input
                  type="text"
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                />
              </FormControl>
              <FormControl>
                <FormLabel>Description</FormLabel>
                <Input
                  type="text"
                  value={description}
                  onChange={(e) => setDescription(e.target.value)}
                />
              </FormControl>
            </Stack>
          }
        />

        <LeftRightLayout
          left={<Text fontSize="lg">Group Owner</Text>}
          right={
            <Stack spacing={4}>
              <Text ml={2}>{owner}</Text>
            </Stack>
          }
        />

        <LeftRightLayout
          left={
            <>
              <Text fontSize="lg">Group Manager (s)</Text>
              <Text color="default.secondary" mt={4} fontSize="sm">
                Can edit the configuration for this group and add/remove
                members.
              </Text>
            </>
          }
          right={
            <>
              <Stack spacing={2}>
                {groupManagers.length === 0 && <Text>No group managers</Text>}
                {groupManagers.map((manager: Member) => (
                  <Flex
                    key={manager.email}
                    align="center"
                    justifyContent="space-between"
                  >
                    <Text ml={2}>{manager.email}</Text>
                    <Button
                      border="1px solid"
                      borderColor="border.neutral.tertiary"
                      size="sm"
                      bg="white"
                      shadow="sm"
                      onClick={async () => {
                        const resp = await axios.delete(
                          `${BACKEND_URL}/api/v1/group-management/groups/${groupId}/members/${manager.email}`,
                          {
                            headers: {
                              Authorization: `Bearer ${auth.user?.access_token}`,
                            },
                          }
                        );

                        if (resp.status > 199 && resp.status < 300) {
                          toast({
                            title: "Member removed",
                            status: "success",
                            duration: 2000,
                            isClosable: true,
                          });
                        } else {
                          toast({
                            title: "Error removing member",
                            status: "error",
                            duration: 2000,
                            isClosable: true,
                          });
                        }

                        navigate(0);
                      }}
                    >
                      Remove
                    </Button>
                  </Flex>
                ))}
              </Stack>

              <Button
                variant="link"
                color="blue.400"
                size="sm"
                mt={4}
                onClick={() => {
                  navigate(`/groups/${groupId}/members`);
                  navigate(0);
                }}
              >
                Add Manager
              </Button>
            </>
          }
        />

        <Box>
          <Text fontSize="lg">Roles</Text>
          <Text color="default.secondary" mt={4} fontSize="sm">
            Choose the roles to assign to members of this group.
          </Text>

          <TableContainer mt={4}>
            <Table variant="simple" size="sm">
              <Thead>
                <Tr>
                  <Th>Role</Th>
                  <Th>Description</Th>
                  <Th />
                </Tr>
              </Thead>
              <Tbody>
                {roles?.map((role, index) => (
                  <Tr key={index}>
                    <Td>
                      <Code colorScheme="gray">{role}</Code>
                    </Td>
                    {/* <Td>{role.description}</Td> */}

                    <Td>
                      <IconButton
                        aria-label="Delete Role"
                        icon={<FiTrash2 />}
                        size="sm"
                        bg=""
                      />
                    </Td>
                  </Tr>
                ))}
              </Tbody>
            </Table>
          </TableContainer>
          <Button variant="link" color="blue.400" size="sm" mt={4}>
            Add Role
          </Button>
        </Box>

        <LeftRightLayout
          left={
            <Text fontSize="lg">Automatically add users to this group</Text>
          }
          right={
            <Flex justifyContent="flex-end">
              <Switch colorScheme="blackAlpha" />
            </Flex>
          }
        />
      </StackedBorderBox>

      <Stack direction="row" mt={8} spacing={4}>
        <ActionButton onClick={handleSaveChanges}>Save Changes</ActionButton>

        <ActionButton
          onClick={transferOwnershipDisclosure.onOpen}
          colorScheme="red"
          isDisabled={!hasOwnerPriv}
        >
          Transfer Ownership
        </ActionButton>

        <TransferOwnershipModal
          isOpen={transferOwnershipDisclosure.isOpen}
          onClose={transferOwnershipDisclosure.onClose}
          groupId={groupId}
          auth={auth}
          navigate={navigate}
          currOwner={owner}
        />

        <Button border="1px solid" borderColor="border.neutral.secondary">
          Archive Group
        </Button>
      </Stack>
    </>
  );
};
