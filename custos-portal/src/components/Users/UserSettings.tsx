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
  Code,
  Spinner,
  HStack,
} from "@chakra-ui/react";
import { PageTitle } from "../PageTitle";
import { ActionButton } from "../ActionButton";
import { Link, useParams } from "react-router-dom";
import { FaArrowLeft } from "react-icons/fa6";
import { LeftRightLayout } from "../LeftRightLayout";
import { FiTrash2 } from "react-icons/fi";
import { StackedBorderBox } from "../StackedBorderBox";
import { BACKEND_URL, CLIENT_ID } from "../../lib/constants";
import { useApi } from "../../hooks/useApi";
import { isEmpty } from "../../lib/util";
import { useEffect, useState } from "react";
import { useAuth } from "react-oidc-context";

// const DUMMY_ROLES: any = [
//   {
//     application: "Grafana",
//     role: "grafana:viewer",
//     description: "Grafana Viewer",
//   },
//   {
//     application: "Grafana",
//     role: "grafana:editor",
//     description: "Grafana Editor",
//   },
//   {
//     application: "Grafana",
//     role: "grafana:admin",
//     description: "Grafana Admin",
//   },
// ];

// const DUMMY_ACTIVITY: any = [
//   {
//     action: "User Created",
//     timestamp: "2021-10-01",
//   },
//   {
//     action: "User Disabled",
//     timestamp: "2021-10-01",
//   },
//   {
//     action: "User Enabled",
//     timestamp: "2021-10-01",
//   },
// ];

export const UserSettings = () => {
  const { email } = useParams();
  const auth = useAuth();

  const [user, setUser] = useState<User | null>(null);
  const [group, setGroup] = useState<any | null>(null);

  useEffect(() => {
    async function fetchData() {
      const userResp = await fetch(
        `${BACKEND_URL}/api/v1/user-management/user/profile/${email}`,
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
      setGroup(groupData);
    }

    fetchData();
  }, []);

  if (!user || !group) {
    return (
      <NavContainer activeTab="Users">
        <Spinner />
      </NavContainer>
    );
  }

  console.log(user);

  return (
    <>
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
          <ActionButton icon={FiTrash2} onClick={() => {}}>
            Disable User
          </ActionButton>
        </Flex>

        <StackedBorderBox>
          <LeftRightLayout
            left={<Text fontSize="lg">Basic Information</Text>}
            right={
              <>
                <Stack spacing={4}>
                  <FormControl color="default.default">
                    <FormLabel>Name</FormLabel>
                    <Input
                      type="text"
                      value={user.first_name + " " + user.last_name}
                    />
                  </FormControl>
                  <FormControl>
                    <FormLabel>Email</FormLabel>
                    <Input type="text" value={user.email} />
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
                    <Th>Actions</Th>
                  </Tr>
                </Thead>
                <Tbody>
                  <Tr>
                    <Td>
                      <Link to="/groups/1">Group 1</Link>
                    </Td>
                    <Td>Admin</Td>
                    <Td>Stella Zhou</Td>
                    <Td>
                      {/* remove icon */}
                      <IconButton
                        aria-label="Delete Role"
                        icon={<FiTrash2 />}
                        size="sm"
                        bg=""
                      />
                    </Td>
                  </Tr>
                </Tbody>
              </Table>
            </TableContainer>
          </Box>

          <Box>
            <Text fontSize="lg">Client Roles</Text>
            {!isEmpty(user.client_roles) ? (
              <>
                <Text mt={2} color="gray.600">
                  Through their group memberships, this user has the following
                  client roles
                </Text>

                <HStack mt={2}>
                  {user.client_roles?.map((role: string) => (
                    <Code key={role} size="lg">
                      {role}
                    </Code>
                  ))}
                </HStack>
              </>
            ) : (
              <Text mt={2} color="gray.600">
                This user has no client roles
              </Text>
            )}
          </Box>

          <Box>
            <Text fontSize="lg">Realm Roles</Text>
            {!isEmpty(user.realm_roles) ? (
              <>
                <Text mt={2} color="gray.600">
                  Through their group memberships, this user has the following
                  realm roles
                </Text>

                <HStack mt={2}>
                  {user.client_roles?.map((role: string) => (
                    <Code key={role} size="lg">
                      {role}
                    </Code>
                  ))}
                </HStack>
              </>
            ) : (
              <Text mt={2} color="gray.600">
                This user has no realm roles
              </Text>
            )}
          </Box>

          {/* <Box>
            <Text fontSize="lg">Activity</Text>
            {
              // eslint-disable-next-line @typescript-eslint/no-explicit-any
              DUMMY_ACTIVITY.map((activity: any) => (
                <Flex key={activity.action} gap={4} mt={4}>
                  <Text color="gray.400">{activity.timestamp}</Text>
                  <Text fontWeight="bold">{activity.action}</Text>
                </Flex>
              ))
            }
          </Box> */}
        </StackedBorderBox>
      </NavContainer>
    </>
  );
};
