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
