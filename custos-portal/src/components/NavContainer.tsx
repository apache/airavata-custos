import React, { useState, useEffect } from 'react';
import {
  Grid,
  GridItem,
  Heading,
  Stack,
  Box,
  Icon,
  Text,
  Flex,
  Button,
  Drawer,
  DrawerOverlay,
  DrawerContent,
  DrawerCloseButton,
  DrawerBody,
  useDisclosure,
} from '@chakra-ui/react';
import { Link } from 'react-router-dom';
import { FiUser, FiUsers, FiChevronLeft, FiChevronRight, FiMenu } from "react-icons/fi";
import { AiOutlineAppstore } from "react-icons/ai";
import { IconType } from "react-icons";
import { MdLogout } from "react-icons/md";
import { useAuth } from 'react-oidc-context';

interface NavContainerProps {
  activeTab: string;
  children: React.ReactNode;
}

interface NavItemProps {
  to: string;
  icon: IconType;
  text: string;
  activeTab: string;
  isCollapsed: boolean;
}

const NavItem = ({ to, icon, text, activeTab, isCollapsed }: NavItemProps) => {
  const isActive = activeTab.toLowerCase() === text.toLowerCase();
  return (
    <Link to={to}>
      <Stack
        direction="row"
        align="center"
        color={isActive ? 'black' : 'default.secondary'}
        py={2}
        px={1}
        _hover={{ bg: 'gray.100' }}
        fontSize="sm"
      >
        <Icon as={icon} />
        {!isCollapsed && <Text fontWeight="semibold">{text}</Text>}
      </Stack>
    </Link>
  );
};

export const NavContainer = ({ activeTab, children }: NavContainerProps) => {
  const auth = useAuth();
  const [isCollapsed, setIsCollapsed] = useState(false);
  const [isMobile, setIsMobile] = useState(window.innerWidth < 768);
  const { isOpen, onOpen, onClose } = useDisclosure();

  useEffect(() => {
    const handleResize = () => setIsMobile(window.innerWidth < 768);
    window.addEventListener('resize', handleResize);
    return () => window.removeEventListener('resize', handleResize);
  }, []);

  if (isMobile) {
    return (
      <>
        <Button onClick={onOpen} variant="ghost" position="fixed" top={4} left={4}>
          <Icon as={FiMenu} w={6} h={6} />
        </Button>
        <Drawer isOpen={isOpen} placement="left" onClose={onClose}>
          <DrawerOverlay />
          <DrawerContent bg="#F7F7F7">
            <DrawerCloseButton />
            <DrawerBody p={4}>
              <Heading size="md">Custos Auth Portal</Heading>
              <Stack direction="column" mt={4}>
                <NavItem to="/applications" icon={AiOutlineAppstore} text="Applications" activeTab={activeTab} isCollapsed={false} />
                <NavItem to="/groups" icon={FiUsers} text="Groups" activeTab={activeTab} isCollapsed={false} />
                <NavItem to="/users" icon={FiUser} text="Users" activeTab={activeTab} isCollapsed={false} />
              </Stack>
              <Box mt="auto">
                <Text fontWeight="bold">{auth.user?.profile?.name}</Text>
                <Text fontSize="sm" color="gray.500">{auth.user?.profile?.email}</Text>
                <Button
                  variant="unstyled"
                  w="fit-content"
                  size="sm"
                  _hover={{ color: "gray.500" }}
                  onClick={async () => {
                    await auth.removeUser();
                    onClose();
                  }}
                >
                  <Flex alignItems="center" gap={2} w="fit-content">
                    <Icon as={MdLogout} />
                    <Text as="span">Logout</Text>
                  </Flex>
                </Button>
              </Box>
            </DrawerBody>
          </DrawerContent>
        </Drawer>
        <Box p={4} pt={20}>
          {children}
        </Box>
      </>
    );
  }

  return (
    <Grid templateColumns="repeat(15, 1fr)" minHeight="100vh">
      <GridItem colSpan={isCollapsed ? 1 : 3} minWidth={isCollapsed ? "60px" : "240px"} maxWidth={isCollapsed ? "60px" : "240px"} bg="#F7F7F7">
        <Flex h="100vh" p={4} direction="column" justifyContent="space-between">
          <Box>
            <Flex justifyContent="space-between" align="center">
              {!isCollapsed && <Heading size="md">Custos Auth Portal</Heading>}
              <Button variant="ghost" onClick={() => setIsCollapsed(!isCollapsed)} size="sm">
                <Icon as={isCollapsed ? FiChevronRight : FiChevronLeft} />
              </Button>
            </Flex>
            <Stack direction="column" mt={4}>
              <NavItem to="/applications" icon={AiOutlineAppstore} text="Applications" activeTab={activeTab} isCollapsed={isCollapsed} />
              <NavItem to="/groups" icon={FiUsers} text="Groups" activeTab={activeTab} isCollapsed={isCollapsed} />
              <NavItem to="/users" icon={FiUser} text="Users" activeTab={activeTab} isCollapsed={isCollapsed} />
            </Stack>
          </Box>
          <Box mt="auto">
            {!isCollapsed && (
              <>
                <Text fontWeight="bold">{auth.user?.profile?.name}</Text>
                <Text fontSize="sm" color="gray.500">{auth.user?.profile?.email}</Text>
              </>
            )}
            <Button
              variant="unstyled"
              w="fit-content"
              size="sm"
              _hover={{ color: "gray.500" }}
              onClick={async () => {
                await auth.removeUser();
              }}
            >
              <Flex alignItems="center" gap={2} w="fit-content">
                <Icon as={MdLogout} />
                {!isCollapsed && <Text as="span">Logout</Text>}
              </Flex>
            </Button>
          </Box>
        </Flex>
      </GridItem>
      <GridItem colSpan={isCollapsed ? 14 : 12} p={16} minWidth="0">
        {children}
      </GridItem>
    </Grid>
  );
};
