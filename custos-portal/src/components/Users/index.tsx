import { useEffect, useRef, useState } from "react";
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
  Button,
  useDisclosure,
  HStack,
  IconButton,
  Select,
} from "@chakra-ui/react";
import { PageTitle } from "../PageTitle";
import { ActionButton } from "../ActionButton";
import { CiSearch } from "react-icons/ci";
import { User } from "../../interfaces/Users";
import { Link } from "react-router-dom";
import { BACKEND_URL } from "../../lib/constants";
import { timeAgo } from "../../lib/util";
import { AddUserModal } from "./AddUserModal";
import { DeleteUserModal } from "./DeleteUserModal";
import { useAuth } from "react-oidc-context";
import { BiPlus } from "react-icons/bi";
import { FcNext, FcPrevious } from "react-icons/fc";

export const Users = () => {
  const addUserModal = useDisclosure();
  const deleteUserModal = useDisclosure();
  const [toDeleteUserId, setToDeleteUserId] = useState<string>("");
  const [offset, setOffset] = useState<number>(0);
  const [limit, setLimit] = useState<number>(10);
  const [pageNumber, setPageNumber] = useState<number>(1);
  const [users, setUsers] = useState<User[]>([]);
  const [length, setLength] = useState<number>(0);
  const [search, setSearch] = useState({
    key: "",
    value: "",
  });
  const auth = useAuth();
  const isFirstRender = useRef(true);

  useEffect(() => {
    async function fetchUsers() {
      const urlSearchParams = new URLSearchParams();
      urlSearchParams.append("offset", offset.toString());
      urlSearchParams.append("limit", limit.toString());

      if (search.key !== "") {
        urlSearchParams.append(search.key, search.value);
      }

      const resp = await fetch(
        `${BACKEND_URL}/api/v1/user-management/users?${urlSearchParams.toString()}`,
        {
          headers: {
            Authorization: `Bearer ${auth.user?.access_token}`,
          },
        }
      );
      const data = await resp.json();
      setLength(data.length);
      setUsers(data.profiles);
    }

    const delayDebounce = setTimeout(fetchUsers, 500);

    if (isFirstRender.current) {
      isFirstRender.current = false;
      fetchUsers();
      return;
    }

    return () => clearTimeout(delayDebounce); // Cleanup timeout on every change
  }, [offset, limit, search]);

  return (
    <>
      <AddUserModal
        isOpen={addUserModal.isOpen}
        onClose={addUserModal.onClose}
        auth={auth}
      />

      <DeleteUserModal
        isOpen={deleteUserModal.isOpen}
        onClose={deleteUserModal.onClose}
        userId={toDeleteUserId}
      />

      <NavContainer activeTab="Users">
        <Flex justifyContent="space-between" alignItems="flex-start">
          <Box>
            <PageTitle>Users</PageTitle>
            <Text color="gray.500" mt={2}>
              View and manage the list of all end users.
            </Text>
          </Box>

          <ActionButton
            icon={BiPlus}
            onClick={() => {
              addUserModal.onOpen();
            }}
          >
            Add User
          </ActionButton>
        </Flex>

        <HStack mt={4}>
          <Select
            value={limit}
            w="200px"
            onChange={(e) => {
              setLimit(parseInt(e.target.value));
              setOffset(0);
              setPageNumber(1);
            }}
          >
            <option value="10">10 per page</option>
            <option value="20">20 per page</option>
            <option value="50">50 per page</option>
            <option value="100">100 per page</option>
          </Select>

          <Select
            value={search.key}
            w="200px"
            onChange={(e) => {
              setSearch({ ...search, key: e.target.value });
            }}
          >
            <option value="">All</option>
            <option value="username">Email</option>
            <option value="first_name">First Name</option>
            <option value="last_name">Last Name</option>
          </Select>

          <InputGroup>
            <InputRightElement pointerEvents="none">
              <Icon as={CiSearch} color="black" />
            </InputRightElement>
            <Input
              value={search.value}
              onChange={(e) => {
                setSearch({ ...search, value: e.target.value });
              }}
              isDisabled={search.key === ""}
              type="text"
              placeholder={
                search.key === ""
                  ? "Showing all users"
                  : `Search by ${search.key}`
              }
              _focus={{
                borderColor: "black",
              }}
              _hover={{
                borderColor: "black",
              }}
            />
          </InputGroup>
        </HStack>

        {/* TABLE */}

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
              {users?.map((user: User) => (
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
                  <Td>{new Date(parseInt(user.created_at)).toDateString()}</Td>
                  <Td>
                    {user.last_modified_at
                      ? timeAgo(new Date(parseInt(user.last_modified_at)))
                      : "N/A"}
                  </Td>
                  <Td>
                    <Button
                      colorScheme="red"
                      size="xs"
                      onClick={() => {
                        setToDeleteUserId(user.email);
                        deleteUserModal.onOpen();
                      }}
                    >
                      Delete
                    </Button>
                  </Td>
                </Tr>
              ))}
            </Tbody>
          </Table>
        </TableContainer>

        <HStack mt={4}>
          <IconButton
            icon={<FcPrevious />}
            bg="transparent"
            aria-label="previous"
            onClick={() => {
              setOffset(offset - limit);
              setPageNumber(pageNumber - 1);
            }}
            isDisabled={offset === 0}
          />
          <Text>
            Showing {offset + 1} -{" "}
            {offset + limit > length ? length : offset + limit}
          </Text>
          <IconButton
            icon={<FcNext />}
            aria-label="next"
            bg="transparent"
            isDisabled={offset + limit >= length}
            onClick={() => {
              setOffset(offset + limit);
              setPageNumber(pageNumber + 1);
            }}
          />
        </HStack>
      </NavContainer>
    </>
  );
};
