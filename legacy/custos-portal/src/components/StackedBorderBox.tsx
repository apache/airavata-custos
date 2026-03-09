import { Divider, Stack } from "@chakra-ui/react";

export const StackedBorderBox = ({
  children,
}: {
  children: React.ReactNode;
}) => {
  return (
    <>
      <Stack
        border="1px solid"
        borderColor="border.neutral.tertiary"
        rounded="xl"
        p={8}
        mt={8}
        divider={<Divider />}
        spacing={8}
      >
        {children}
      </Stack>
    </>
  );
};
