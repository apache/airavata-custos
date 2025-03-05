import { NavContainer } from "../NavContainer";
import {
  Box,
  Flex,
  Text,
  Input,
  Icon,
  TableContainer,
  Table,
  Thead,
  Tr,
  Th,
  Td,
  Tbody,
  Stack,
  FormControl,
  FormLabel,
  IconButton,
  Spinner,
  HStack,
  Button,
  useToast,
  Code,
  useDisclosure,
} from "@chakra-ui/react";

import { PageTitle } from "../PageTitle";
import { Link, useParams } from "react-router-dom";
import { FaArrowLeft } from "react-icons/fa6";
import { LeftRightLayout } from "../LeftRightLayout";
import { FiTrash2 } from "react-icons/fi";
import { StackedBorderBox } from "../StackedBorderBox";
import { BACKEND_URL, CLIENT_ID } from "../../lib/constants";
import { useEffect, useState } from "react";
import { useAuth } from "react-oidc-context";
import { MetadataAttribute, User } from "../../interfaces/Users";
import { Group } from "../../interfaces/Groups";
import { IoAdd } from "react-icons/io5";
import { AttributeValue } from "./AttributeValue";
import { AddAttributeBox } from "./AddAttributeBox";
import { RolesEditor } from "./RolesEditor";
import { DeleteUserModal } from "./DeleteUserModal";
import { UserTrailsTable } from "./UserTrailsTable";

export const UserSettings = () => {
  const { email } = useParams();
  const auth = useAuth();

  const [user, setUser] = useState<User | null>(null);
  const [groups, setGroups] = useState<Group[] | null>(null);
  const [addingAttribute, setAddingAttribute] = useState(false);
  const [clientRoles, setClientRoles] = useState<string[]>([]);
  const [realmRoles, setRealmRoles] = useState<string[]>([]);
  const deleteUserModal = useDisclosure();
  const [userTrails, setUserTrails] = useState<MetadataAttribute[]>([]);
  const [updatedName, setUpdatedName] = useState({
    first_name: "",
    last_name: "",
  });

  const toast = useToast();

  useEffect(() => {
    async function fetchUserTrails(): Promise<void> {
      const resp = await fetch(
        `${BACKEND_URL}/api/v1/user-management/users/${email}/metadata-trails`,
        {
          headers: {
            Authorization: `Bearer ${auth.user?.access_token}`,
          },
        }
      );

      if (resp.ok) {
        const data = await resp.json();
        setUserTrails(data.attribute_audit);
      }
    }

    fetchUserTrails();
  }, [clientRoles, realmRoles]);

  useEffect(() => {
    async function fetchData() {
      const userResp = await fetch(
        `${BACKEND_URL}/api/v1/user-management/users/${email}`,
        {
          headers: {
            Authorization: `Bearer ${auth.user?.access_token}`,
          },
        }
      );
      const userData = await userResp.json();

      const groupResp = await fetch(
        `${BACKEND_URL}/api/v1/group-management/users/${email}/group-memberships`,
        {
          headers: {
            client_id: CLIENT_ID,
            userId: email,
            Authorization: `Bearer ${auth.user?.access_token}`,
          },
        }
      );
      const groupData = await groupResp.json();

      setUser(userData);
      setUpdatedName({
        first_name: userData.first_name,
        last_name: userData.last_name,
      });
      setClientRoles(userData.client_roles ?? []);
      setRealmRoles(userData.realm_roles ?? []);
      if (!groupData.groups) {
        setGroups([]);
      } else {
        setGroups(groupData.groups);
      }
    }

    fetchData();
  }, []);

  if (!user || !groups) {
    return (
      <NavContainer activeTab="Users">
        <Spinner />
      </NavContainer>
    );
  }

  return (
    <>
      <DeleteUserModal
        isOpen={deleteUserModal.isOpen}
        onClose={deleteUserModal.onClose}
        userId={user.username}
      />
      <NavContainer activeTab="Users">
        <Link to="/users">
          <Flex alignItems="center" gap={2} color="default.secondary">
            <Icon as={FaArrowLeft} />
            <Text fontWeight="bold" fontSize="sm">
              Back to Users
            </Text>
          </Flex>
        </Link>

        <Flex mt={4} justify="space-between">
          <Box>
            <PageTitle>
              {user.first_name} {user.last_name}
            </PageTitle>
            <Text color="default.secondary" mt={2}>
              {user.email}
            </Text>
          </Box>
          <Button
            leftIcon={<Icon as={FiTrash2} />}
            colorScheme="red"
            variant="outline"
            onClick={deleteUserModal.onOpen}
          >
            Delete User
          </Button>
        </Flex>

        <StackedBorderBox>
          <LeftRightLayout
            left={<Text fontSize="lg">Basic Information</Text>}
            right={
              <>
                <Stack spacing={4}>
                  <FormControl color="default.default">
                    <FormLabel>First Name</FormLabel>
                    <Input
                      type="text"
                      value={updatedName.first_name}
                      onChange={(e) =>
                        setUpdatedName((prev) => ({
                          ...prev,
                          first_name: e.target.value,
                        }))
                      }
                    />
                  </FormControl>
                  <FormControl color="default.default">
                    <FormLabel>Last Name</FormLabel>
                    <Input
                      type="text"
                      value={updatedName.last_name}
                      onChange={(e) =>
                        setUpdatedName((prev) => ({
                          ...prev,
                          last_name: e.target.value,
                        }))
                      }
                    />
                  </FormControl>
                  <FormControl>
                    <FormLabel>Email</FormLabel>
                    <Input type="text" value={user.email} disabled={true} />
                  </FormControl>
                  <FormControl>
                    <FormLabel>Joined</FormLabel>
                    <Input
                      type="text"
                      disabled={true}
                      value={new Date(
                        parseInt(user.created_at)
                      ).toLocaleString()}
                    />
                  </FormControl>
                  <FormControl>
                    <FormLabel>Last Modified</FormLabel>
                    <Input
                      type="text"
                      disabled={true}
                      value={new Date(
                        parseInt(user.last_modified_at)
                      ).toLocaleString()}
                    />
                  </FormControl>

                  {updatedName.first_name !== user.first_name ||
                  updatedName.last_name !== user.last_name ? (
                    <HStack justify="flex-start">
                      <Button
                        onClick={async () => {
                          // @ts-expect-error this is fine
                          setUser((prev) => ({
                            ...prev,
                            first_name: updatedName.first_name,
                            last_name: updatedName.last_name,
                          }));

                          const resp = await fetch(
                            `${BACKEND_URL}/api/v1/user-management/users/${email}`,
                            {
                              method: "PATCH",
                              headers: {
                                "Content-Type": "application/json",
                                Authorization: `Bearer ${auth.user?.access_token}`,
                              },
                              body: JSON.stringify({
                                firstName: updatedName.first_name,
                                lastName: updatedName.last_name,
                              }),
                            }
                          );

                          if (resp.ok) {
                            toast({
                              title: "Information saved!",
                              description: "User information has been updated",
                              status: "success",
                              duration: 3000,
                              isClosable: true,
                            });
                          } else {
                            toast({
                              title: "An error occurred",
                              description: "Please try again later",
                              status: "error",
                              duration: 3000,
                              isClosable: true,
                            });
                          }
                        }}
                      >
                        Save
                      </Button>
                      <Button
                        onClick={() => {
                          setUpdatedName({
                            first_name: user.first_name,
                            last_name: user.last_name,
                          });
                        }}
                      >
                        Cancel
                      </Button>
                    </HStack>
                  ) : (
                    <></>
                  )}
                </Stack>
              </>
            }
          />

          <Box>
            <Text fontSize="lg">Group Memberships</Text>
            <TableContainer mt={4}>
              <Table variant="simple">
                <Thead>
                  <Tr>
                    <Th>Name</Th>
                    <Th>Role</Th>
                    <Th>Owner</Th>
                    <Th>Client Roles</Th>
                    <Th>Realm Roles</Th>
                  </Tr>
                </Thead>
                <Tbody>
                  {groups?.map((group: Group) => {
                    return (
                      <Tr key={group.id}>
                        <Td>
                          <Link to={`/groups/${group.id}/settings`}>
                            {group.name}
                          </Link>
                        </Td>
                        <Td>{group.requester_role}</Td>
                        <Td>{group.owner_id}</Td>
                        <Td>
                          {group?.client_roles?.map((role) => (
                            <Text key={role}>
                              <Code>{role}</Code>
                            </Text>
                          ))}
                        </Td>
                        <Td>
                          {group?.realm_roles?.map((role) => (
                            <Text key={role}>
                              <Code>{role}</Code>
                            </Text>
                          ))}
                        </Td>
                      </Tr>
                    );
                  })}
                </Tbody>
              </Table>
            </TableContainer>
          </Box>

          <Box>
            <Text fontSize="lg">Client Roles</Text>
            <Text my={2} color="gray.600">
              User specific client roles. To view group inherited roles, see the
              Group Memberships section.
            </Text>

            <RolesEditor
              username={user.username}
              roles={clientRoles}
              setRoles={setClientRoles}
              type="client"
            />
          </Box>

          <Box>
            <Text fontSize="lg">Realm Roles</Text>
            <Text my={2} color="gray.600">
              User specific realm roles. To view group inherited roles, see the
              Group Memberships section.
            </Text>

            <RolesEditor
              username={user.username}
              roles={realmRoles}
              setRoles={setRealmRoles}
              type="realm"
            />
          </Box>
          <Box>
            <Text fontSize="lg">User Attributes</Text>
            <Text my={2} color="gray.600">
              User attributes are specific to users and are not inherited from
              group memberships. To delete an attribute, delete all values for
              it. Attributes are keys that can have multiple values.
            </Text>
            <HStack flexWrap="wrap">
              {user?.attributes?.map((attr) => (
                <Box key={attr.key}>
                  {attr.values.map((value) => (
                    <AttributeValue
                      key={value}
                      username={user.username}
                      attrKey={attr.key}
                      value={value}
                    />
                  ))}
                </Box>
              ))}

              {addingAttribute ? (
                <AddAttributeBox
                  username={user.username}
                  setAddingAttribute={setAddingAttribute}
                />
              ) : (
                <IconButton
                  icon={<IoAdd />}
                  size="sm"
                  onClick={() => setAddingAttribute((a) => !a)}
                />
              )}
            </HStack>
          </Box>

          <Box>
            <Text fontSize="lg">User Trails</Text>

            {userTrails.length > 0 ? (
              <UserTrailsTable trails={userTrails} />
            ) : (
              <Text color="gray.500" mt={2}>
                No user trails found
              </Text>
            )}
          </Box>
        </StackedBorderBox>
      </NavContainer>
    </>
  );
};
