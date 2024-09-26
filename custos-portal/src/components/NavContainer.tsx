import {
  Grid,
  GridItem,
  Heading,
  Stack,
  Box,
  Icon,
  Text,
  Flex,
  Button
} from '@chakra-ui/react';
import { Link } from 'react-router-dom';
import { FiUser, FiUsers } from "react-icons/fi";
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
}

const NavItem = ({ to, icon, text, activeTab }: NavItemProps) => {
  const isActive = activeTab.toLowerCase() === text.toLowerCase();
  return (
    <Link to={to}>
      <Stack
        direction='row'
        align='center'
        color={isActive ? 'black' : 'default.secondary'}
        py={2}
        px={1}
        _hover={{
          bg: 'gray.100',
        }}
        fontSize='sm'
      >
        <Icon as={icon} />
        <Text fontWeight='semibold'>{text}</Text>
      </Stack>
    </Link>
  )
}


export const NavContainer = ({ activeTab, children }: NavContainerProps) => {
  const auth = useAuth();
  return (
    <>
      <Grid templateColumns='repeat(15, 1fr)'>
        <GridItem colSpan={3} bg='#F7F7F7'>
          <Flex h='100vh' >
            <Box position='fixed'>
              <Flex justifyContent='space-between' direction="column" h='100vh' p={4} >
              <Box>
                <Heading size='md'>
                  Custos Auth Portal
                </Heading>

                <Stack direction='column'  mt={4}>
                  <NavItem to='/applications' icon={AiOutlineAppstore} text="Applications" activeTab={activeTab} />
                  <NavItem to='/groups' icon={FiUsers} text="Groups" activeTab={activeTab} />
                  <NavItem to='/users' icon={FiUser} text="Users" activeTab={activeTab} />
                </Stack>
              </Box>

              <Box>
                <Text fontWeight='bold'>
                  {auth.user?.profile?.name}
                </Text>
                <Text fontSize='sm' color='gray.500'>
                  {auth.user?.profile?.email}
                </Text>
                <Button 
                  variant='unstyled' 
                  w='fit-content'
                  size='sm'
                  _hover={{
                    color: 'gray.500'
                  }}
                  onClick={async () => {
                    await auth.removeUser();
                  }}
                >
                  <Flex alignItems='center' gap={2} w='fit-content'>
                      <Icon as={MdLogout} />
                      <Text as='span' >Logout</Text>
                  </Flex>
                </Button>
              </Box>
            </Flex>
            </Box>
          </Flex>

        </GridItem>
        <GridItem colSpan={12} p={16}>
          {children}
        </GridItem>
      </Grid>
    </>
  )
}

