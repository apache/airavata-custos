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
  Spinner,
} from "@chakra-ui/react";
import { PageTitle } from "../PageTitle";
import { ActionButton } from "../ActionButton";
import { CiSearch } from "react-icons/ci";
import { User } from "../../interfaces/Users";
import { Link } from "react-router-dom";
import { useApi } from "../../hooks/useApi";
import { BACKEND_URL } from "../../lib/constants";
import { timeAgo } from "../../lib/util";

export const Users = () => {
  const offset = 0;
  const limit = 10;


  const urlSearchParams = new URLSearchParams();
  urlSearchParams.append("offset", offset.toString());
  urlSearchParams.append("limit", limit.toString());

  const allUsers = useApi(
    `${BACKEND_URL}/api/v1/user-management/users/profile?${urlSearchParams.toString()}`
  );
    
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
        {allUsers?.isPending ? (
          <Box textAlign="center" mt={4}>
            <Spinner />
          </Box>
        ) : (
          <TableContainer mt={4}>
            <Table variant="simple">
              <Thead>
                <Tr>
                  <Th>Name</Th>
                  <Th>Email</Th>
                  <Th>Joined</Th>
                  <Th>Last Modified</Th>
                  <Th>Actions</Th>
                </Tr>
              </Thead>

              <Tbody>
                {allUsers?.data?.profiles?.map((user: User) => (
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
                          {user.first_name} {user.last_name}
                        </Text>
                      </Link>
                    </Td>
                    <Td>{user.email}</Td>
                    <Td>
                      {new Date(parseInt(user.created_at)).toDateString()}
                    </Td>
                    <Td>
                      {user.last_modified_at
                        ? timeAgo(new Date(parseInt(user.last_modified_at)))
                        : "N/A"}
                    </Td>
                    <Td>
                      <ActionButton onClick={() => {}}>Disable</ActionButton>
                    </Td>
                  </Tr>
                ))}
              </Tbody>
            </Table>
          </TableContainer>
        )}
      </NavContainer>
    </>
  );
};
