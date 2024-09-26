/* eslint-disable @typescript-eslint/no-explicit-any */
import {
  Modal,
  ModalOverlay,
  ModalContent,
  ModalHeader,
  ModalFooter,
  ModalBody,
  ModalCloseButton,
  Button,
  FormControl,
  FormLabel,
  Input,
  FormHelperText,
  Flex,
  useToast,
} from "@chakra-ui/react";
import axios, { AxiosResponse } from "axios";
import { useState } from "react";
import { BACKEND_URL } from "../../lib/constants";

export const TransferOwnershipModal = ({
  isOpen,
  onClose,
  groupId,
  auth,
  navigate,
  currOwner,
}: {
  isOpen: boolean;
  onClose: () => void;
  groupId: string;
  auth: any;
  navigate: any;
  currOwner: string;
}) => {
  const [email, setEmail] = useState("");
  const toast = useToast();

  const handleTransferOwnership = async () => {
    console.log("transferring owner to...", email);
    const emailRegex = /^[a-zA-Z0-9._-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,6}$/;

    if (!emailRegex.test(email)) {
      toast({
        title: "Invalid email",
        status: "error",
        duration: 2000,
        isClosable: true,
      });
      return;
    }

    // find the current owner

    // first, add the email as an owner
    const resp = (await axios
      .post(
        `${BACKEND_URL}/api/v1/group-management/groups/${groupId}/members`,
        {
          username: email,
          type: "OWNER",
        },
        {
          headers: {
            Authorization: `Bearer ${auth.user?.access_token}`,
          },
        }
      )
      .catch((error) => {
        toast({
          title: "Error transferring ownership",
          description: error.response.data.error,
          status: "error",
          duration: 2000,
          isClosable: true,
        });
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
      })) as AxiosResponse<any, any>;

    console.log(resp.status);

    if (resp.status > 199 && resp.status < 300) {
      await handleRemoveMember(currOwner);
      navigate(0);
    }
  };

  const handleRemoveMember = async (email: string) => {
    try {
      await axios.delete(
        `${BACKEND_URL}/api/v1/group-management/groups/${groupId}/members/${email}`,
        {
          headers: {
            Authorization: `Bearer ${auth.user?.access_token}`,
          },
        }
      );
    } catch (error) {
      toast({
        title: "Error removing member",
        status: "error",
        description: (error as any)?.response?.data,
        duration: 4000,
        isClosable: true,
      });
      return;
    }
  };

  return (
    <>
      <Modal isOpen={isOpen} onClose={onClose}>
        <ModalOverlay />
        <ModalContent>
          <ModalHeader>Transfer Ownership</ModalHeader>
          <ModalCloseButton />
          <ModalBody>
            <FormControl>
              <FormLabel>New Owner Email</FormLabel>
              <Flex gap={2}>
                <Input
                  type="email"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                />
                <Button colorScheme="red" onClick={handleTransferOwnership}>
                  Transfer
                </Button>
              </Flex>

              <FormHelperText>
                Transferring ownership for GroupId: {groupId}.
              </FormHelperText>
            </FormControl>
          </ModalBody>

          <ModalFooter />
        </ModalContent>
      </Modal>
    </>
  );
};
