/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements. See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership. The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License. You may obtain a copy of the License at
 *
 *  http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing,
 *  software distributed under the License is distributed on an
 *  "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 *  KIND, either express or implied. See the License for the
 *  specific language governing permissions and limitations
 *  under the License.
 */

import React, { useState, useEffect, memo } from "react";
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
  Spacer,
} from "@chakra-ui/react";
import { Link } from "react-router-dom";
import {
  FiUser,
  FiUsers,
  FiChevronLeft,
  FiChevronRight,
  FiMenu,
} from "react-icons/fi";
import { AiOutlineAppstore } from "react-icons/ai";
import { IconType } from "react-icons";
import { MdLogout } from "react-icons/md";
import { useAuth } from "react-oidc-context";

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
  onClose: () => void;
}

const NavItem = memo(
  ({ to, icon, text, activeTab, isCollapsed, onClose }: NavItemProps) => {
    const isActive = activeTab.toLowerCase() === text.toLowerCase();
    return (
      <Link to={to} onClick={onClose}>
        <Stack
          direction="row"
          align="center"
          color={isActive ? "black" : "default.secondary"}
          py={2}
          px={1}
          _hover={{ bg: "gray.100" }}
          fontSize="sm"
        >
          <Icon as={icon} />
          {!isCollapsed && <Text fontWeight="semibold">{text}</Text>}
        </Stack>
      </Link>
    );
  }
);

export const NavContainer = memo(
  ({ activeTab, children }: NavContainerProps) => {
    const auth = useAuth();

    const [isCollapsed, setIsCollapsed] = useState(() => {
      const saved = localStorage.getItem("navCollapsed");
      return saved ? JSON.parse(saved) : false;
    });

    const [isMobile, setIsMobile] = useState(window.innerWidth < 768);
    const { isOpen, onOpen, onClose } = useDisclosure();

    useEffect(() => {
      const handleResize = () => setIsMobile(window.innerWidth < 768);
      window.addEventListener("resize", handleResize);
      return () => window.removeEventListener("resize", handleResize);
    }, []);

    useEffect(() => {
      localStorage.setItem("navCollapsed", JSON.stringify(isCollapsed));
    }, [isCollapsed]);

    const toggleCollapse = () => {
      setIsCollapsed((prev: boolean) => !prev);
    };

    if (isMobile) {
      return (
        <>
          <Box position="fixed" top={4} left={4} zIndex={10}>
            <Button onClick={onOpen} variant="ghost">
              <Icon as={FiMenu} w={6} h={6} />
            </Button>
          </Box>
          <Drawer isOpen={isOpen} placement="left" onClose={onClose}>
            <DrawerOverlay />
            <DrawerContent bg="#F7F7F7">
              <DrawerCloseButton />
              <DrawerBody p={4} display="flex" flexDirection="column">
                <Heading size="md">Custos Auth Portal</Heading>
                <Stack direction="column" mt={4}>
                  <NavItem
                    to="/applications"
                    icon={AiOutlineAppstore}
                    text="Applications"
                    activeTab={activeTab}
                    isCollapsed={false}
                    onClose={onClose}
                  />
                  <NavItem
                    to="/groups"
                    icon={FiUsers}
                    text="Groups"
                    activeTab={activeTab}
                    isCollapsed={false}
                    onClose={onClose}
                  />
                  <NavItem
                    to="/users"
                    icon={FiUser}
                    text="Users"
                    activeTab={activeTab}
                    isCollapsed={false}
                    onClose={onClose}
                  />
                </Stack>
                <Spacer />
                <Box>
                  <Text fontWeight="bold">{auth.user?.profile?.name}</Text>
                  <Text fontSize="sm" color="gray.500">
                    {auth.user?.profile?.email}
                  </Text>
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
          <Box p={5} pt={16}>
            {children}
          </Box>
        </>
      );
    }

    return (
      <Grid templateColumns="repeat(15, 1fr)" minHeight="100vh">
        <GridItem
          colSpan={isCollapsed ? 1 : 3}
          minWidth={isCollapsed ? "60px" : "240px"}
          maxWidth={isCollapsed ? "60px" : "240px"}
          bg="#F7F7F7"
          position="fixed"
          h="100vh"
        >
          <Flex
            h="100vh"
            p={4}
            direction="column"
            justifyContent="space-between"
          >
            <Box>
              <Flex justifyContent="space-between" align="center">
                {!isCollapsed && <Heading size="md">Custos Portal</Heading>}
                <Button variant="ghost" onClick={toggleCollapse} size="sm">
                  <Icon as={isCollapsed ? FiChevronRight : FiChevronLeft} />
                </Button>
              </Flex>
              <Stack direction="column" mt={4}>
                <NavItem
                  to="/applications"
                  icon={AiOutlineAppstore}
                  text="Applications"
                  activeTab={activeTab}
                  isCollapsed={isCollapsed}
                  onClose={onClose}
                />
                <NavItem
                  to="/groups"
                  icon={FiUsers}
                  text="Groups"
                  activeTab={activeTab}
                  isCollapsed={isCollapsed}
                  onClose={onClose}
                />
                <NavItem
                  to="/users"
                  icon={FiUser}
                  text="Users"
                  activeTab={activeTab}
                  isCollapsed={isCollapsed}
                  onClose={onClose}
                />
              </Stack>
            </Box>
            <Box mt="auto">
              {!isCollapsed && (
                <>
                  <Text fontWeight="bold">{auth.user?.profile?.name}</Text>
                  <Text fontSize="sm" color="gray.500">
                    {auth.user?.profile?.email}
                  </Text>
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
        <GridItem
          colSpan={isCollapsed ? 14 : 15}
          p={10}
          ml={isCollapsed ? "60px" : "240px"}
          minWidth="0"
        >
          {children}
        </GridItem>
      </Grid>
    );
  }
);
