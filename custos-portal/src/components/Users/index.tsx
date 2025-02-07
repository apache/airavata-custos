import { NavContainer } from "../NavContainer";
import {
  Box,
  Flex,
  Text,
  Input,
  InputGroup,
  InputRightElement,
  Icon,
  TableContainer,
  Table,
  Thead,
  Tr,
  Th,
  Td,
  Tbody,
} from "@chakra-ui/react";
import { PageTitle } from "../PageTitle";
import { ActionButton } from "../ActionButton";
import { CiSearch } from "react-icons/ci";
import { User } from "../../interfaces/Users";
import { Link } from "react-router-dom";

const DUMMY_DATA: User[] = [
  {
    name: "Stella Zhou",
    email: "xz106@au.edu",
    joined: "2021-10-01",
    lastSignedIn: "2021-10-01",
  },
  {
    name: "John Doe",
    email: "asdf@asdf.edu",
    joined: "2021-10-01",
    lastSignedIn: "2021-10-01",
  },
  {
    name: "Jane Doe",
    email: "jane.doe@asdf.edu",
    joined: "2021-10-01",
    lastSignedIn: "2021-10-01",
  },
];

export const Users = () => {
  return (
    <>
      <NavContainer activeTab="Users">
        <Flex justifyContent="space-between" alignItems="flex-start">
          <Box>
            <PageTitle>Users</PageTitle>
            <Text color="gray.500" mt={2}>
              View and manage the list of all end users.
            </Text>
          </Box>
        </Flex>

        <InputGroup mt={4}>
          <InputRightElement pointerEvents="none">
            <Icon as={CiSearch} color="black" />
          </InputRightElement>
          <Input
            type="text"
            placeholder="Search users"
            _focus={{
              borderColor: "black",
            }}
            _hover={{
              borderColor: "black",
            }}
          />
        </InputGroup>

        {/* TABLE */}
        <TableContainer mt={4}>
          <Table variant="simple">
            <Thead>
              <Tr>
                <Th>Name</Th>
                <Th>Email</Th>
                <Th>Joined</Th>
                <Th>Last Signed In</Th>
                <Th>Actions</Th>
              </Tr>
            </Thead>

            <Tbody>
              {DUMMY_DATA.map((user) => (
                <Tr key={user.email}>
                  <Td>
                    <Link to={`/users/${user.email}`}>
                      <Text
                        color="blue.400"
                        _hover={{
                          color: "blue.600",
                          cursor: "pointer",
                        }}
                      >
                        {user.name}
                      </Text>
                    </Link>
                  </Td>
                  <Td>{user.email}</Td>
                  <Td>{user.joined}</Td>
                  <Td>{user.lastSignedIn}</Td>
                  <Td>
                    <ActionButton onClick={() => {}}>Disable</ActionButton>
                  </Td>
                </Tr>
              ))}
            </Tbody>
          </Table>
        </TableContainer>
      </NavContainer>
    </>
  );
};
