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
