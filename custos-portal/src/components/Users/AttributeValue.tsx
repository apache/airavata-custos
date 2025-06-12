import { Code, IconButton, HStack, useToast } from "@chakra-ui/react";
import { IoClose } from "react-icons/io5";
import { useState } from "react";
import { BACKEND_URL } from "../../lib/constants";
import { useNavigate } from "react-router-dom";
import { useAuth } from "react-oidc-context";

export const AttributeValue = ({
  username,
  attrKey,
  value,
}: {
  username: string;
  attrKey: string;
  value: string;
}) => {
  const [hovering, setHovering] = useState(false);
  const navigate = useNavigate();
  const toast = useToast();
  const auth = useAuth();

  const handleDelete = async () => {
    const resp = await fetch(
      `${BACKEND_URL}/api/v1/user-management/users/${username}/attributes`,
      {
        method: "DELETE",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${auth.user?.access_token}`,
        },
        body: JSON.stringify({
          attributes: [
            {
              key: attrKey,
              values: [value],
            },
          ],
        }),
      }
    );

    if (resp.ok) {
      toast({
        title: "Attribute removed",
        description: "The attribute has been removed successfully",
        status: "success",
        duration: 3000,
        isClosable: true,
      });
      navigate(0);
    } else {
      toast({
        title: "An error occurred",
        description: "Please try again later",
        status: "error",
        duration: 3000,
        isClosable: true,
      });
    }
  };

  return (
    <HStack
      onMouseEnter={() => setHovering(true)}
      onMouseLeave={() => setHovering(false)}
      bg={hovering ? "red.100" : "gray.50"}
      px={2}
      py={1}
      borderRadius="md"
      spacing={2}
      align="center"
      justify="space-between"
    >
      <Code fontSize="sm" px={2} py={1} borderRadius="md" bg="transparent">
        {attrKey} | {value}
      </Code>

      {hovering && (
        <IconButton
          aria-label="Remove attribute"
          icon={<IoClose />}
          size="sm"
          variant="ghost"
          color="gray.600"
          _hover={{ color: "red.500" }}
          onClick={handleDelete}
        />
      )}
    </HStack>
  );
};
