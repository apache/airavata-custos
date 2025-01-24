import { SimpleGrid, Box } from "@chakra-ui/react";

export const LeftRightLayout = ({
  left,
  right,
}: {
  left: React.ReactNode;
  right: React.ReactNode;
}) => {
  return (
    <SimpleGrid columns={2} spacing={8}>
      <Box>{left}</Box>
      <Box>{right}</Box>
    </SimpleGrid>
  );
};
