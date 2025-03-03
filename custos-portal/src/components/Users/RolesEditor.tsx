import { Button, HStack, useToast } from "@chakra-ui/react";
import { useState } from "react";
import { TagsInput } from "react-tag-input-component";
import { BACKEND_URL } from "../../lib/constants";
import { useAuth } from "react-oidc-context";

export const RolesEditor = ({
  username,
  roles,
  setRoles,
  type,
}: {
  username: string;
  roles: string[];
  setRoles: (roles: string[]) => void;
  type: string; // should be `client` or `realm`
}) => {
  const [roleChanges, setRoleChanges] = useState({
    added: [] as string[],
    removed: [] as string[],
  });
  const auth = useAuth();
  const toast = useToast();

  const handleSaveChanges = async () => {
    console.log("Saving changes...");
    console.log("Added roles:", roleChanges.added);
    console.log("Removed roles:", roleChanges.removed);

    let req1Status = true;
    let req2Status = true;

    if (roleChanges.added.length > 0) {
      const addResp = await fetch(
        `${BACKEND_URL}/api/v1/user-management/roles`,
        {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
            Authorization: `Bearer ${auth.user?.access_token}`,
          },
          body: JSON.stringify({
            roles: roleChanges.added,
            usernames: [username],
            roleType: type,
          }),
        }
      );

      req1Status = addResp.ok;
    }

    if (roleChanges.removed.length > 0) {
      const removeResp = await fetch(
        `${BACKEND_URL}/api/v1/user-management/roles`,
        {
          method: "DELETE",
          headers: {
            "Content-Type": "application/json",
            Authorization: `Bearer ${auth.user?.access_token}`,
          },
          body: JSON.stringify({
            roles: roleChanges.removed,
            usernames: [username],
            roleType: type,
          }),
        }
      );

      req2Status = removeResp.ok;
    }

    if (req1Status && req2Status) {
      toast({
        title: "Roles updated",
        description: "The roles have been updated successfully",
        status: "success",
        duration: 3000,
        isClosable: true,
      });

      // Update `roles` after successful save
      setRoles([
        ...roles.filter((role) => !roleChanges.removed.includes(role)), // Remove deleted roles
        ...roleChanges.added, // Add new roles
      ]);
    } else {
      toast({
        title: "An error occurred",
        description: "Please try again later",
        status: "error",
        duration: 3000,
        isClosable: true,
      });
    }

    setRoleChanges({ added: [], removed: [] });
  };

  console.log("roleChanges", roleChanges);

  return (
    <>
      <TagsInput
        value={
          roles
            .filter((role) => !roleChanges.removed.includes(role)) // Show current roles excluding removed ones
            .concat(roleChanges.added) // Show added roles
        }
        onChange={(updatedRoles) => {
          setRoleChanges({
            added: updatedRoles.filter((role) => !roles.includes(role)), // Roles that were newly added
            removed: roles.filter((role) => !updatedRoles.includes(role)), // Roles that were removed
          });
        }}
      />
      {roleChanges.added.length > 0 || roleChanges.removed.length > 0 ? (
        <HStack mt={2}>
          <Button
            variant="outline"
            onClick={() => setRoleChanges({ added: [], removed: [] })}
          >
            Cancel
          </Button>
          <Button onClick={handleSaveChanges} bg="black" color="white">
            Save Changes
          </Button>
        </HStack>
      ) : (
        <></>
      )}
    </>
  );
};
