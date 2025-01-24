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
} from "@chakra-ui/react";
import { PageTitle } from "../PageTitle";
import { ActionButton } from "../ActionButton";
import { Link, useParams } from "react-router-dom";
import { FaArrowLeft } from "react-icons/fa6";
import { LeftRightLayout } from "../LeftRightLayout";
import { FiTrash2 } from "react-icons/fi";
import { StackedBorderBox } from "../StackedBorderBox";

const DUMMY_ROLES: any = [
  {
    application: "Grafana",
    role: "grafana:viewer",
    description: "Grafana Viewer",
  },
  {
    application: "Grafana",
    role: "grafana:editor",
    description: "Grafana Editor",
  },
  {
    application: "Grafana",
    role: "grafana:admin",
    description: "Grafana Admin",
  },
];

const DUMMY_ACTIVITY: any = [
  {
    action: "User Created",
    timestamp: "2021-10-01",
  },
  {
    action: "User Disabled",
    timestamp: "2021-10-01",
  },
  {
    action: "User Enabled",
    timestamp: "2021-10-01",
  },
];

export const UserSettings = () => {
  const { email } = useParams();

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
            <PageTitle>John Doe</PageTitle>
            <Text color="default.secondary" mt={2}>
              {email}
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
                    <Input type="text" />
                  </FormControl>
                  <FormControl>
                    <FormLabel>Email</FormLabel>
                    <Input type="text" />
                  </FormControl>
                  <FormControl>
                    <FormLabel>Joined</FormLabel>
                    <Input type="text" disabled={true} />
                  </FormControl>
                  <FormControl>
                    <FormLabel>Last Signed In</FormLabel>
                    <Input type="text" disabled={true} />
                  </FormControl>
                </Stack>
              </>
            }
          />

          <Box>
            <Text fontSize="lg">Groups</Text>
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
            <Text fontSize="lg">Roles</Text>
            <Text mt={2} color="gray.600">
              Through their group memberships, this user has the following roles
            </Text>

            <TableContainer mt={4}>
              <Table variant="simple">
                <Thead>
                  <Tr>
                    <Th>Application</Th>
                    <Th>Role</Th>
                    <Th>Description</Th>
                  </Tr>
                </Thead>
                <Tbody>
                  {DUMMY_ROLES.map((role: any) => (
                    <Tr key={role.role}>
                      <Td>{role.application}</Td>
                      <Td>
                        <Code>{role.role}</Code>
                      </Td>
                      <Td>{role.description}</Td>
                    </Tr>
                  ))}
                </Tbody>
              </Table>
            </TableContainer>
          </Box>

          <Box>
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
          </Box>
        </StackedBorderBox>
      </NavContainer>
    </>
  );
};
