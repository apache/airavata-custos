import { Center, Img, SimpleGrid, Box, Flex, Text, Avatar, AvatarGroup, Button } from "@chakra-ui/react"
import cuate from "../../public/cuate.png"
import airavata from "../../public/airavata.png"
import otherLogo from "../../public/otherLogo.svg"
import { PageTitle } from "../PageTitle"
import { useAuth } from "react-oidc-context"

export const Login = () => {
  const auth = useAuth();
  return (
    <>
      <Center height='100vh'>
        <SimpleGrid columns={{
          base: 1,
          lg: 2
        }} spacing={32} alignItems='center'>
          <Img 
            src={cuate} 
            alt="cuate" 
            maxW='400px' 
            display={{
              base: 'none',
              lg: 'block'
            }}
          />

          <Box>
            <Flex gap={4} alignItems='center'>
                <AvatarGroup size='md' max={2}>
                  <Avatar name='Airavata' src={airavata} />
                  <Avatar name='Other' src={otherLogo} />
                </AvatarGroup>

                <Box>
                  <Text fontWeight='bold'>
                    Custos Auth Portal
                  </Text>
                  <Text color='gray.500' fontSize='sm'>
                    Developed as a part of the VEDA project
                  </Text>
                </Box>
            </Flex>

            <Box my={8}>
              <PageTitle>Welcome</PageTitle>
            </Box>

            <Button
              bg='black'
              color='white'
              _hover={{ bg: 'gray.800' }}
              _active={{ bg: 'gray.900' }}
              rounded='full'
              w='300px'
              onClick={() => {
                console.log('Sign in clicked')
                auth.signinRedirect()
              }}
            >
              Institution Login
            </Button>
          </Box>
        </SimpleGrid>
      </Center>
    </>
  )
}
