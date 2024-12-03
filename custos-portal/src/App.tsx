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

import { Routes, Route, BrowserRouter } from 'react-router-dom';
import { Heading } from '@chakra-ui/react';
import { Groups } from './components/Groups';
import { NavContainer } from './components/NavContainer';
import { GroupDetails } from './components/Groups/GroupDetails';
import { Login } from './components/Login';
import ProtectedComponent from './components/ProtectedComponent';

function NotImplemented() {
  return (
    <NavContainer activeTab='N/A'>
      <Heading size='lg' fontWeight={500}>
        Not Implemented
      </Heading>
    </NavContainer>
  );
}


export default function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<Login />} />
          <Route path="/applications" element={<ProtectedComponent Component={NotImplemented}  />} />
          <Route path="/users" element={<ProtectedComponent Component={NotImplemented}  />} />
          <Route path="/groups/:id/:path" element={<ProtectedComponent Component={GroupDetails}  />} />
          <Route path="/groups" element={<ProtectedComponent Component={Groups}  />} />
      </Routes>
    </BrowserRouter>
  );
}
