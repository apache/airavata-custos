/* eslint-disable @typescript-eslint/no-unused-vars */
import {
  FormControl,
  Input,
  FormLabel,
  Modal,
  ModalBody,
  ModalCloseButton,
  ModalContent,
  ModalHeader,
  ModalOverlay,
  Text,
  SimpleGrid,
  ModalFooter,
  VStack,
  Button,
  Divider,
  FormHelperText,
  useToast,
  Spinner,
  IconButton,
  Flex,
} from "@chakra-ui/react";
import { useState } from "react";
import { AuthContextProps } from "react-oidc-context";
import { TagsInput } from "react-tag-input-component";
import { BACKEND_URL } from "../../lib/constants";
import { useNavigate } from "react-router-dom";
import { Attribute } from "../../interfaces/Users";
import { RegisterRequest } from "../../interfaces/Users";
import { FaTrash } from "react-icons/fa6";

export const AddUserModal = ({
  isOpen,
  onClose,
  auth,
}: {
  isOpen: boolean;
  onClose: () => void;
  auth: AuthContextProps;
}) => {
  const [registerRequest, setRegisterRequest] = useState({
    username: "",
    email: "",
    firstName: "",
    lastName: "",
    status: "ACTIVE",
    attributes: [],
    clientRoles: [],
    realmRoles: [],
    lastModifiedAt: Date.now(),
    type: "END_USER",
  } as RegisterRequest);
  const [loading, setLoading] = useState(false);
  const toast = useToast();
  const navigate = useNavigate();

  const registerUser = async () => {
    setLoading(true);

    if (!registerRequest.username || !registerRequest.email) {
      toast({
        title: "Please fill out all required fields",
        description: "First name, last name, and email are required",
        status: "error",
        duration: 3000,
        isClosable: true,
      });
      setLoading(false);
      return;
    }
    if (registerRequest.attributes.length !== 0) {
      for (const attr of registerRequest.attributes) {
        if (!attr.key || attr.values.length === 0) {
          toast({
            title: "Please fill out all required fields",
            description: "Attribute key and values are required",
            status: "error",
            duration: 3000,
            isClosable: true,
          });
          setLoading(false);
          return;
        }
      }
    }

    const resp = await fetch(`${BACKEND_URL}/api/v1/user-management/user`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${auth.user?.access_token}`,
      },
      body: JSON.stringify(registerRequest),
    });

    if (resp.ok) {
      toast({
        title: "User created",
        description: "The user has been created successfully",
        status: "success",
        duration: 3000,
        isClosable: true,
      });
      navigate(0);
      onClose();
    } else {
      toast({
        title: "An error occurred",
        description: "Please try again later",
        status: "error",
        duration: 3000,
        isClosable: true,
      });
    }

    setLoading(false);
  };

  return (
    <>
      <Modal isOpen={isOpen} onClose={onClose} size="lg">
        <ModalOverlay />
        <ModalContent>
          <ModalHeader>Add Tenant User</ModalHeader>
          <ModalCloseButton />
          <ModalBody>
            <VStack spacing={4} divider={<Divider />}>
              <SimpleGrid columns={2} spacing={4} width="100%">
                <FormControl>
                  <FormLabel>First Name</FormLabel>
                  <Input
                    placeholder="Jane"
                    value={registerRequest.firstName}
                    onChange={(e) =>
                      setRegisterRequest((prev) => ({
                        ...prev,
                        firstName: e.target.value,
                      }))
                    }
                  />
                </FormControl>
                <FormControl>
                  <FormLabel>Last Name</FormLabel>
                  <Input
                    placeholder="Doe"
                    value={registerRequest.lastName}
                    onChange={(e) =>
                      setRegisterRequest((prev) => ({
                        ...prev,
                        lastName: e.target.value,
                      }))
                    }
                  />
                </FormControl>
              </SimpleGrid>

              <FormControl>
                <FormLabel>Email</FormLabel>
                <Input
                  placeholder="jane.doe@gmail.com"
                  value={registerRequest.email}
                  onChange={(e) =>
                    setRegisterRequest((prev) => ({
                      ...prev,
                      username: e.target.value,
                      email: e.target.value,
                    }))
                  }
                />
              </FormControl>

              {/* create a way to add attributes. the format of each attribute is listed above. make sure there's a plus button to add more */}
              <FormControl>
                <FormLabel>Attributes</FormLabel>
                <FormHelperText>
                  Add custom attributes for the user. Attribute values is a list
                  (press enter to add a new value).
                </FormHelperText>

                {registerRequest.attributes.map((attr, index) => {
                  return (
                    <Flex key={index} alignItems="end" gap={4}>
                      <SimpleGrid
                        columns={2}
                        spacing={4}
                        mt={2}
                        alignItems="center"
                      >
                        {/* Key Field */}
                        <FormControl flexGrow={1}>
                          <FormLabel>Key</FormLabel>
                          <Input
                            placeholder="key"
                            value={attr.key}
                            onChange={(e) => {
                              const newAttributes = [
                                ...registerRequest.attributes,
                              ];
                              newAttributes[index].key = e.target.value;
                              setRegisterRequest((prev) => ({
                                ...prev,
                                attributes: newAttributes,
                              }));
                            }}
                          />
                        </FormControl>

                        {/* Values Field */}
                        <FormControl flexGrow={2}>
                          <FormLabel>Values</FormLabel>
                          <TagsInput
                            value={attr.values}
                            onChange={(values) => {
                              const newAttributes = [
                                ...registerRequest.attributes,
                              ];
                              newAttributes[index].values = values;
                              setRegisterRequest((prev) => ({
                                ...prev,
                                attributes: newAttributes,
                              }));
                            }}
                          />
                        </FormControl>
                      </SimpleGrid>

                      {/* Trash Button - Minimal Space */}
                      <IconButton
                        aria-label="Remove attribute"
                        icon={<FaTrash />}
                        size="sm"
                        p={2}
                        colorScheme="red"
                        onClick={() => {
                          const newAttributes = [...registerRequest.attributes];
                          newAttributes.splice(index, 1);
                          setRegisterRequest((prev) => ({
                            ...prev,
                            attributes: newAttributes,
                          }));
                        }}
                      />
                    </Flex>
                  );
                })}
                <Button
                  mt={2}
                  alignSelf="flex-start"
                  onClick={() => {
                    setRegisterRequest((prev) => ({
                      ...prev,
                      attributes: [
                        ...prev.attributes,
                        {
                          key: "",
                          values: [],
                        },
                      ],
                    }));
                  }}
                  size="sm"
                >
                  Add Attribute
                </Button>
              </FormControl>

              <FormControl>
                <FormLabel>Client Roles</FormLabel>
                <FormHelperText mb={2}>
                  Add a list of specific client roles for the user. Client roles
                  can also be inherited from groups.
                </FormHelperText>
                <TagsInput
                  value={registerRequest.clientRoles}
                  onChange={(values) =>
                    setRegisterRequest((prev) => ({
                      ...prev,
                      clientRoles: values,
                    }))
                  }
                />
              </FormControl>

              <FormControl>
                <FormLabel>Realm Roles</FormLabel>
                <FormHelperText mb={2}>
                  Add a list of specific realm roles for the user. Realm roles
                  can also be inherited from groups.
                </FormHelperText>
                <TagsInput
                  value={registerRequest.realmRoles}
                  onChange={(values) =>
                    setRegisterRequest((prev) => ({
                      ...prev,
                      realmRoles: values,
                    }))
                  }
                />
              </FormControl>

              <Button
                onClick={registerUser}
                bg="black"
                color="white"
                width="100%"
                isDisabled={loading}
              >
                {loading ? <Spinner /> : "Add User"}
              </Button>
            </VStack>
          </ModalBody>

          <ModalFooter />
        </ModalContent>
      </Modal>
    </>
  );
};
