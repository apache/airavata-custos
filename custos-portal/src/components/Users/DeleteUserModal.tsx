import { useState } from "react";
import {
  Modal,
  ModalOverlay,
  ModalContent,
  ModalHeader,
  ModalCloseButton,
  ModalBody,
  ModalFooter,
  Button,
  Text,
  Spinner,
} from "@chakra-ui/react";
import { useAuth } from "react-oidc-context";
import { BACKEND_URL } from "../../lib/constants";
import { useNavigate } from "react-router-dom";

export const DeleteUserModal = ({
  userId,
  isOpen,
  onClose,
}: {
  userId: string;
  isOpen: boolean;
  onClose: () => void;
}) => {
  const [loading, setLoading] = useState(false);
  const auth = useAuth();
  const navigate = useNavigate();

  const deleteUser = async () => {
    setLoading(true);
    const response = await fetch(
      `${BACKEND_URL}/api/v1/user-management/users/${userId}`,
      {
        method: "DELETE",
        headers: {
          Authorization: `Bearer ${auth.user?.access_token}`,
        },
      }
    );

    if (response.ok) {
      onClose();
      navigate("/users");
    }
    setLoading(false);
  };

  return (
    <Modal isOpen={isOpen} onClose={onClose}>
      <ModalOverlay />
      <ModalContent>
        <ModalHeader>Delete User</ModalHeader>
        <ModalCloseButton />
        <ModalBody>
          <Text>
            Are you sure you want to delete user{" "}
            <Text as="span" fontWeight="bold">
              {userId}
            </Text>
            ?
          </Text>
        </ModalBody>
        <ModalFooter>
          <Button colorScheme="red" onClick={deleteUser} width="100%">
            {loading ? <Spinner size="sm" /> : "Delete"}
          </Button>
        </ModalFooter>
      </ModalContent>
    </Modal>
  );
};
