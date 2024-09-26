import { Heading } from "@chakra-ui/react";

interface PageTitleProps {
  children: React.ReactNode;
  size?: string;
}

export const PageTitle = ({ children, size }: PageTitleProps) => {
  return (
    <Heading size={size} color='default.default' fontWeight={500}>
      {children}
    </Heading>
  )
};
