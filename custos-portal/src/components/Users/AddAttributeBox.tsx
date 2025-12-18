import { Button, HStack, Input, useToast } from "@chakra-ui/react";
import { useState } from "react";
import { useAuth } from "react-oidc-context";
import { BACKEND_URL } from "../../lib/constants";
import { useNavigate } from "react-router-dom";

export const AddAttributeBox = ({ username }: { username: string }) => {
  const [key, setKey] = useState<string>("");
  const [value, setValue] = useState<string>("");
  const auth = useAuth();
  const toast = useToast();
  const navigate = useNavigate();

  const handleAdd = async () => {
    const resp = await fetch(
      `${BACKEND_URL}/api/v1/user-management/users/${username}/attributes`,
      {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${auth.user?.access_token}`,
        },
        body: JSON.stringify({
          attributes: [
            {
              key,
              values: [value],
            },
          ],
        }),
      }
    );

    if (resp.ok) {
      setKey("");
      setValue("");
      toast({
        title: "Attribute added",
        description: "The attribute has been added successfully",
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
    <HStack>
      <Input
        placeholder="Key"
        value={key}
        size="sm"
        onChange={(e) => setKey(e.target.value)}
      />
      <Input
        placeholder="Value"
        value={value}
        size="sm"
        onChange={(e) => setValue(e.target.value)}
      />
      <Button size="sm" bg="black" color="white" onClick={handleAdd}>
        Add
      </Button>
    </HStack>
  );
};
